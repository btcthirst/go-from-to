package ui

import (
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/reports"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewReportScreen(state *AppState) fyne.CanvasObject {
	// Вибір формату
	formatSelect := widget.NewSelect([]string{"XLSX", "ODS"}, func(s string) {})
	formatSelect.SetSelected("XLSX")

	// Вибір шаблону (опціонально)
	templateLabel := widget.NewLabel("Шаблон: не обрано")
	templateBtn := widget.NewButton("Обрати шаблон...", func() {
		// file dialog для вибору template.xlsx
	})

	// Фільтр дат
	fromEntry := widget.NewEntry()
	fromEntry.SetPlaceHolder("від (дд.мм.рррр)")
	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("до (дд.мм.рррр)")

	// Чекбокси секцій звіту
	includeSummary := widget.NewCheck("Підсумок", func(b bool) {})
	includeByCategory := widget.NewCheck("За категоріями", func(b bool) {})
	includeMonthly := widget.NewCheck("По місяцях", func(b bool) {})
	includeChart := widget.NewCheck("Графік", func(b bool) {})
	includeSummary.SetChecked(true)
	includeByCategory.SetChecked(true)
	includeMonthly.SetChecked(true)
	includeChart.SetChecked(true)

	// Прогрес та генерація
	progress := widget.NewProgressBar()
	progress.Hide()

	generateBtn := widget.NewButtonWithIcon("Згенерувати звіт", theme.DocumentSaveIcon(), func() {
		d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if writer == nil {
				return
			}
			progress.Show()
			go func() {
				reporter := &reports.XLSXReporter{}
				report := buildReport(state)
				err := reporter.Generate(report, writer.URI().Path())
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				} else {
					dialog.ShowInformation("Готово", "Звіт успішно збережено!", fyne.CurrentApp().Driver().AllWindows()[0])
				}
			}()
		}, fyne.CurrentApp().Driver().AllWindows()[0])
		d.SetFileName("звіт.xlsx")
		d.Show()
	})

	return container.NewVBox(
		widget.NewCard("Налаштування", "",
			container.NewVBox(
				container.NewHBox(widget.NewLabel("Формат:"), formatSelect),
				container.NewHBox(templateLabel, templateBtn),
				container.NewHBox(widget.NewLabel("Період:"), fromEntry, widget.NewLabel("—"), toEntry),
			),
		),
		widget.NewCard("Секції звіту", "",
			container.NewVBox(includeSummary, includeByCategory, includeMonthly, includeChart),
		),
		progress,
		generateBtn,
	)
}

func buildReport(state *AppState) *models.Report {
	// Тут можна застосувати фільтри за датами та іншими параметрами
	return &models.Report{
		Transactions: state.Transactions,
	}
}
