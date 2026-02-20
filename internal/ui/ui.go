// Package ui надає UI на основі Fyne для вибору файлів та обробки даних
package ui

import (
	"excel-parser/internal/app"
	"excel-parser/internal/excel/writer"
	"excel-parser/internal/model"
	"fmt"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// UIManager управляє UI застосунку
type UIManager struct {
	accrualFile string
	paymentFile string
	accruals    []model.AccrualRecord
	results     []model.ResultRecord
	hasResults  bool
	saveBtn     *widget.Button
}

// NewUIManager створює новий UI менеджер
func NewUIManager() *UIManager {
	return &UIManager{
		accrualFile: "",
		paymentFile: "",
		accruals:    []model.AccrualRecord{},
		results:     []model.ResultRecord{},
		hasResults:  false,
		saveBtn:     nil,
	}
}

// Run запускає UI застосунку
func (um *UIManager) Run() {
	fyneAppInstance := fyneapp.NewWithID("excel-parser-app")
	w := fyneAppInstance.NewWindow("Excel/ODS File Processor")
	w.SetTitle("Обробник Excel/ODS файлів")

	// Заголовок
	title := canvas.NewText("Обробник Excel/ODS файлів", nil)
	title.TextSize = 24
	title.TextStyle.Bold = true

	// Лейбли для показу вибраних файлів
	accrualLabel := widget.NewLabel("Файл нарахувань: не вибрано")
	paymentLabel := widget.NewLabel("Файл виписок: не вибрано")
	statusLabel := widget.NewLabel("")

	// Таблиця нарахувань
	accrualsTable := widget.NewTable(
		func() (int, int) {
			if um.accruals == nil {
				return 0, 0
			}
			return len(um.accruals) + 1, 2 // +1 for header, 2 columns: FullName, Account
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				// Header row
				headers := []string{"ПІБ", "Рахунок"}
				label.SetText(headers[id.Col])
				label.TextStyle.Bold = true
			} else {
				idx := id.Row - 1
				if idx < len(um.accruals) {
					accrual := um.accruals[idx]
					switch id.Col {
					case 0:
						label.SetText(accrual.FullName)
					case 1:
						label.SetText(accrual.Account)
					}
				}
			}
		},
	)
	accrualsTable.SetColumnWidth(0, 250)
	accrualsTable.SetColumnWidth(1, 150)

	// Обгортка таблиці нарахувань у скрол
	accrualsScroll := container.NewScroll(accrualsTable)

	// Таблиця результатів платежів
	resultsTable := widget.NewTable(
		func() (int, int) {
			if !um.hasResults {
				return 0, 0
			}
			// Рахуємо тільки звичайні платежі (без відрахувань)
			count := 1 // header
			for _, r := range um.results {
				if !r.IsDeduction {
					count++
				}
			}
			return count, 12 // +1 for header, 12 columns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				// Header row
				headers := []string{
					"№п",
					"Постачальник",
					"Дата",
					"Дт рах.311 Сума",
					"Оборот по Дт",
					"313",
					"63",
					"641",
					"641.1",
					"651",
					"94",
					"Оборот по Кт",
				}
				label.SetText(headers[id.Col])
				label.TextStyle.Bold = true
			} else {
				// Знаходимо індекс результату пропускаючи відрахування
				resultIdx := 0
				displayIdx := 0
				for resultIdx < len(um.results) {
					if !um.results[resultIdx].IsDeduction {
						if displayIdx == id.Row-1 {
							break
						}
						displayIdx++
					}
					resultIdx++
				}

				if resultIdx < len(um.results) {
					result := um.results[resultIdx]
					switch id.Col {
					case 0:
						label.SetText(fmt.Sprintf("%d", id.Row))
					case 1:
						label.SetText(result.Name)
					case 2:
						label.SetText(result.Date)
					case 3:
						label.SetText(fmt.Sprintf("%.2f", result.Sum))
					case 4:
						label.SetText(result.Account)
					case 5, 6, 7, 8, 9, 10:
						label.SetText("")
					case 11:
						label.SetText(result.Counterparty)
					}
				}
			}
		},
	)
	resultsTable.SetColumnWidth(0, 40)
	resultsTable.SetColumnWidth(1, 150)
	resultsTable.SetColumnWidth(2, 100)
	resultsTable.SetColumnWidth(3, 120)
	resultsTable.SetColumnWidth(4, 100)
	for i := 5; i <= 10; i++ {
		resultsTable.SetColumnWidth(i, 60)
	}
	resultsTable.SetColumnWidth(11, 100)

	// Обгортка таблиці результатів у скрол
	resultsScroll := container.NewScroll(resultsTable)

	// Таблиця відрахувань (негативних платежів)
	deductionsTable := widget.NewTable(
		func() (int, int) {
			if !um.hasResults {
				return 0, 0
			}
			// Рахуємо тільки відрахування (негативні платежі)
			count := 1 // header
			for _, r := range um.results {
				if r.IsDeduction {
					count++
				}
			}
			return count, 6 // +1 for header, 6 columns: Date, Sum, ПІБ, Рахунок, Призначення, Контрагент
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				// Header row
				headers := []string{"Дата", "Сума", "ПІБ", "Рахунок", "Призначення", "Контрагент"}
				label.SetText(headers[id.Col])
				label.TextStyle.Bold = true
			} else {
				// Знаходимо індекс результату тільки серед відрахувань
				resultIdx := 0
				displayIdx := 0
				for resultIdx < len(um.results) {
					if um.results[resultIdx].IsDeduction {
						if displayIdx == id.Row-1 {
							break
						}
						displayIdx++
					}
					resultIdx++
				}

				if resultIdx < len(um.results) {
					result := um.results[resultIdx]
					switch id.Col {
					case 0:
						label.SetText(result.Date)
					case 1:
						// Показуємо суму з мінусом як вона була
						label.SetText(fmt.Sprintf("%.2f", result.Sum))
					case 2:
						label.SetText(result.Name)
					case 3:
						label.SetText(result.Account)
					case 4:
						label.SetText(result.Purpose)
					case 5:
						label.SetText(result.Counterparty)
					}
				}
			}
		},
	)
	deductionsTable.SetColumnWidth(0, 100)
	deductionsTable.SetColumnWidth(1, 100)
	deductionsTable.SetColumnWidth(2, 150)
	deductionsTable.SetColumnWidth(3, 120)
	deductionsTable.SetColumnWidth(4, 200)
	deductionsTable.SetColumnWidth(5, 150)

	// Обгортка таблиці відрахувань у скрол
	deductionsScroll := container.NewScroll(deductionsTable)

	// Таби з результатами
	tabs := container.NewAppTabs(
		container.NewTabItem("Нарахування", accrualsScroll),
		container.NewTabItem("Платежі", resultsScroll),
		container.NewTabItem("Відрахування", deductionsScroll),
	)

	accrualBtn := widget.NewButton("Вибрати файл нарахувань (ODS)", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			um.accrualFile = reader.URI().Path()
			accrualLabel.SetText("Файл нарахувань: " + reader.URI().Name())
			statusLabel.SetText("")

			// Одразу обробляємо нарахування у окремій горутині
			go func() {
				accruals, err := app.GetAccruals(um.accrualFile)
				if err != nil {
					fyne.Do(func() {
						statusLabel.SetText("Помилка при читанні нарахувань: " + err.Error())
					})
					return
				}

				fyne.Do(func() {
					um.accruals = accruals
					statusLabel.SetText(fmt.Sprintf("Нарахування завантажені. Записів: %d", len(accruals)))
					accrualsTable.Refresh()
				})
			}()

			// Перевіряємо, чи готові для повної обробки
			um.processIfReady(statusLabel, accrualsTable, resultsTable, deductionsTable)
		}, w)
		fd.Show()
	})

	// Кнопка вибору файлу виписок
	paymentBtn := widget.NewButton("Вибрати файл виписок (XLSX)", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			um.paymentFile = reader.URI().Path()
			paymentLabel.SetText("Файл виписок: " + reader.URI().Name())
			statusLabel.SetText("")
			um.processIfReady(statusLabel, accrualsTable, resultsTable, deductionsTable)
		}, w)
		fd.Show()
	})

	// Кнопка "Зберегти як"
	saveBtn := widget.NewButton("Зберегти як...", func() {
		if !um.hasResults {
			dialog.ShowInformation("Помилка", "Немає результатів для збереження", w)
			return
		}

		fd := dialog.NewFileSave(func(writerCloser fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writerCloser == nil {
				return
			}
			defer writerCloser.Close()

			outputPath := writerCloser.URI().Path()
			if err := writer.WriteResults(um.results, outputPath); err != nil {
				dialog.ShowError(err, w)
				statusLabel.SetText("Помилка при збереженні файлу")
				return
			}
			statusLabel.SetText("Файл успішно збережено: " + outputPath)
			dialog.ShowInformation("Успіх", fmt.Sprintf("Результати збережені у файл:\n%s", outputPath), w)
		}, w)
		fd.SetFileName("summery.ods")
		fd.Show()
	})
	saveBtn.Disable()

	// Контейнер з кнопками вибору файлів
	buttonsBox := container.NewVBox(
		accrualBtn,
		paymentBtn,
	)

	// Контейнер з інформацією про файли
	fileInfoBox := container.NewVBox(
		accrualLabel,
		paymentLabel,
	)

	// Верхня частина UI (файли)
	topBox := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel("Виберіть файли:"),
		buttonsBox,
		widget.NewSeparator(),
		fileInfoBox,
		widget.NewSeparator(),
	)

	// Нижня частина UI (кнопки)
	bottomBox := container.NewVBox(
		widget.NewSeparator(),
		saveBtn,
		statusLabel,
	)

	// Основний контейнер з бордером для кращого розподілу простору
	mainBox := container.NewBorder(
		topBox,    // top
		bottomBox, // bottom
		nil,       // left
		nil,       // right
		tabs,      // center - таблиці займають більшість простору
	)

	w.SetContent(mainBox)

	// Панель UI менеджера
	um.saveBtn = saveBtn

	w.ShowAndRun()
}

// processIfReady обробляє файли, якщо обидва вибрані
func (um *UIManager) processIfReady(statusLabel *widget.Label, accrualsTable, resultsTable, deductionsTable *widget.Table) {
	if um.accrualFile == "" || um.paymentFile == "" {
		return
	}

	fyne.Do(func() {
		statusLabel.SetText("Обробка... будь ласка, зачекайте")
	})

	// Запускаємо обробку в окремій горутині
	go func() {
		// Читаємо нарахування
		accruals, err := app.GetAccruals(um.accrualFile)
		if err != nil {
			fyne.Do(func() {
				statusLabel.SetText("Помилка при читанні нарахувань: " + err.Error())
			})
			return
		}

		// Обробляємо платежі
		results, err := app.ProcessPayments(um.accrualFile, um.paymentFile)
		if err != nil {
			fyne.Do(func() {
				statusLabel.SetText("Помилка при обробці: " + err.Error())
			})
			return
		}

		// Всі UI обновлення повинні бути на головному потоці
		fyne.Do(func() {
			um.accruals = accruals
			um.results = results
			um.hasResults = true
			um.saveBtn.Enable()

			// Рахуємо платежі та відрахування
			paymentCount := 0
			deductionCount := 0
			for _, r := range results {
				if r.IsDeduction {
					deductionCount++
				} else {
					paymentCount++
				}
			}

			statusLabel.SetText(fmt.Sprintf("Обробка завершена. Нарахувань: %d, Платежів: %d, Відрахувань: %d", len(accruals), paymentCount, deductionCount))
			accrualsTable.Refresh()
			resultsTable.Refresh()
			deductionsTable.Refresh()
		})
	}()
}
