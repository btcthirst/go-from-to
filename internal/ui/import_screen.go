package ui

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"bank-analyzer/internal/models"
	"bank-analyzer/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// loadedFile зберігає метадані про один імпортований файл.
type loadedFile struct {
	name       string // коротка назва файлу
	path       string // повний шлях
	bankName   string // назва банку від парсера
	txCount    int    // кількість імпортованих транзакцій
	importedAt time.Time
	err        error // помилка парсингу, якщо була
}

func (f *loadedFile) statusText() string {
	if f.err != nil {
		return "Помилка"
	}
	return fmt.Sprintf("%s · %d транзакцій", f.bankName, f.txCount)
}

// NewImportScreen повертає екран імпорту файлів.
func NewImportScreen(state *AppState) fyne.CanvasObject {
	var mu sync.RWMutex
	var files []loadedFile

	win := fyne.CurrentApp().Driver().AllWindows()[0]

	// --- Зведення (summary) ---
	summaryLabel := widget.NewLabel("Файлів не завантажено")
	summaryLabel.Alignment = fyne.TextAlignCenter

	// --- Список файлів ---
	fileList := widget.NewList(
		func() int {
			return len(files)
		},
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			name := widget.NewLabel("")
			name.TextStyle = fyne.TextStyle{Bold: true}
			status := widget.NewLabel("")
			status.Importance = widget.LowImportance
			return container.NewHBox(icon, container.NewVBox(name, status))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(files) {
				return
			}
			f := files[id]
			box := item.(*fyne.Container)
			inner := box.Objects[1].(*fyne.Container)

			nameLabel := inner.Objects[0].(*widget.Label)
			statusLabel := inner.Objects[1].(*widget.Label)

			nameLabel.SetText(f.name)
			statusLabel.SetText(f.statusText())

			if f.err != nil {
				statusLabel.Importance = widget.DangerImportance
			} else {
				statusLabel.Importance = widget.LowImportance
			}
			statusLabel.Refresh()
		},
	)

	// refreshSummary оновлює підпис зведення під списком.
	refreshSummary := func() {
		total := len(state.Transactions)
		fileCount := len(files)
		switch fileCount {
		case 0:
			summaryLabel.SetText("Файлів не завантажено")
		default:
			summaryLabel.SetText(
				fmt.Sprintf("Завантажено файлів: %d  ·  Транзакцій всього: %d", fileCount, total),
			)
		}
	}

	// importFile — єдина функція імпорту, яку використовують і кнопка, і dropper.
	// Виконується в горутині, UI оновлює через fyne.Do.
	importFile := func(path string) {
		name := filepath.Base(path)

		// Показуємо "в процесі" одразу
		pending := loadedFile{
			name:       name,
			path:       path,
			bankName:   "завантаження...",
			importedAt: time.Now(),
		}
		fyne.Do(func() {
			mu.Lock()
			files = append(files, pending)
			mu.Unlock()
			fileList.Refresh()
		})

		// Парсинг у фоні
		parser, detectErr := state.Registry.Detect(path)

		var txs []*models.Transaction
		var parseErr error

		if detectErr != nil {
			parseErr = detectErr
		} else {
			txs, parseErr = parser.Parse(path)
		}

		// Категоризація — теж поза UI-потоком
		if parseErr == nil {
			state.Categorizer.CategorizeAll(txs)
			state.Resolver.ResolveAll(txs)
		}

		// Оновлення стану і UI — тільки в головному потоці
		fyne.Do(func() {
			mu.Lock()
			defer mu.Unlock()
			// Знайти і оновити pending-запис
			for i := range files {
				if files[i].path == path && files[i].bankName == "завантаження..." {
					if parseErr != nil {
						files[i].err = parseErr
						files[i].bankName = "—"
					} else {
						files[i].bankName = parser.Name()
						files[i].txCount = len(txs)
						state.AppendTransactions(txs)
					}
					break
				}
			}

			fileList.Refresh()
			refreshSummary()
			state.NotifyTransactionsChanged()

			if parseErr != nil {
				dialog.ShowError(
					fmt.Errorf("не вдалося розпізнати файл «%s»:\n%w", name, parseErr),
					win,
				)
			}
		})
	}

	// --- Кнопка вибору файлу ---
	addBtn := widget.NewButtonWithIcon("Додати файл", theme.FolderOpenIcon(), func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if reader == nil {
				return
			}
			path := reader.URI().Path()
			reader.Close()
			go importFile(path)
		}, win)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".csv", ".xlsx", ".pdf", ".ods"}))
		d.Show()
	})

	// --- Кнопка очищення ---
	clearBtn := widget.NewButtonWithIcon("Очистити все", theme.DeleteIcon(), func() {
		dialog.ShowConfirm(
			"Очистити список",
			"Видалити всі завантажені файли та транзакції?",
			func(ok bool) {
				if !ok {
					return
				}

				files = nil
				state.ClearTransactions()

				fileList.Refresh()
				refreshSummary()
				state.NotifyTransactionsChanged()
			},
			win,
		)
	})
	clearBtn.Importance = widget.DangerImportance

	// --- Drag & Drop зона ---
	dropper := widgets.NewFileDropper(func(uris []fyne.URI) {
		for _, u := range uris {
			go importFile(u.Path())
		}
	})

	// ДОДАНО: Перехоплюємо подію скидання файлу у самого вікна.
	win.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		// Отримуємо абсолютні координати нашого віджета
		dPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(dropper)
		dSize := dropper.Size()

		// Перевіряємо, чи в момент Drop мишка знаходилася безпосередньо над нашим віджетом
		if pos.X >= dPos.X && pos.X <= dPos.X+dSize.Width &&
			pos.Y >= dPos.Y && pos.Y <= dPos.Y+dSize.Height {
			dropper.DroppedFiles(uris) // Передаємо файли віджету
		}
	})

	// --- Складання layout ---
	toolbar := container.NewBorder(nil, nil, nil,
		container.NewHBox(addBtn, clearBtn),
	)

	top := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
		dropper,
		widget.NewSeparator(),
	)

	return container.NewBorder(top, summaryLabel, nil, nil, fileList)
}
