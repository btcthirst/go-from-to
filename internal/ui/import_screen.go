package ui

import (
	"bank-analyzer/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewImportScreen(state *AppState) fyne.CanvasObject {
	// Список завантажених файлів
	fileList := widget.NewList(
		func() int { return len(state.loadedFiles) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel(""),
				widget.NewLabel(""), // статус/банк
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// заповнити рядок списку
		},
	)

	// Кнопка вибору файлу
	addBtn := widget.NewButtonWithIcon("Додати файл", theme.FolderOpenIcon(), func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader == nil {
				return
			}
			go func() {
				txs, parseErr := state.Registry.ParseFile(reader.URI().Path())
				if parseErr != nil {
					dialog.ShowError(parseErr, fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}
				state.Categorizer.CategorizeAll(txs)
				state.Transactions = append(state.Transactions, txs...)
				fileList.Refresh()
			}()
		}, fyne.CurrentApp().Driver().AllWindows()[0])
		d.SetFilter(storage.NewExtensionFileFilter([]string{".csv", ".xlsx", ".pdf", ".ods"}))
		d.Show()
	})

	// Drag & Drop зона (кастомний віджет)
	dropper := widgets.NewFileDropper(func(paths []string) {
		for _, path := range paths {
			go func(p string) {
				txs, err := state.Registry.ParseFile(p)
				if err == nil {
					state.Categorizer.CategorizeAll(txs)
					state.Transactions = append(state.Transactions, txs...)
					fileList.Refresh()
				}
			}(path)
		}
	})

	// Зведення після імпорту
	summary := widget.NewRichText()
	// Оновлюється після кожного імпорту

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Перетягніть файли або оберіть вручну"),
			dropper,
			addBtn,
		),
		summary,
		nil, nil,
		fileList,
	)
}
