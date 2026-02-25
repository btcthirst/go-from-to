package ui

import (
	"errors"
	"fmt"
	"time"

	"bank-analyzer/internal/models"
	"bank-analyzer/internal/reports"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/shopspring/decimal"
)

const dateLayout = "02.01.2006"

// reportOptions зберігає всі налаштування генерації звіту.
type reportOptions struct {
	templatePath    string
	from, to        time.Time
	includeSummary  bool
	includeCategory bool
	includeMonthly  bool
	includeChart    bool
}

// NewReportScreen повертає екран генерації звітів.
func NewReportScreen(state *AppState) fyne.CanvasObject {
	win := fyne.CurrentApp().Driver().AllWindows()[0]

	opts := reportOptions{
		includeSummary:  true,
		includeCategory: true,
		includeMonthly:  true,
		includeChart:    true,
	}

	// --- Формат ---
	formatSelect := widget.NewSelect([]string{"XLSX", "ODS"}, nil)
	formatSelect.SetSelected("XLSX")

	// --- Шаблон ---
	templateLabel := widget.NewLabel("не обрано")
	templateLabel.Importance = widget.LowImportance

	templateBtn := widget.NewButtonWithIcon("Обрати шаблон...", theme.FolderOpenIcon(), func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()
			opts.templatePath = reader.URI().Path()
			templateLabel.SetText(reader.URI().Name())
			templateLabel.Importance = widget.MediumImportance
			templateLabel.Refresh()
		}, win)
		d.SetFilter(newFormatFilter(formatSelect.Selected))
		d.Show()
	})

	clearTemplateBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		opts.templatePath = ""
		templateLabel.SetText("не обрано")
		templateLabel.Importance = widget.LowImportance
		templateLabel.Refresh()
	})

	// --- Фільтр дат ---
	fromEntry := widget.NewEntry()
	fromEntry.SetPlaceHolder("дд.мм.рррр")

	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("дд.мм.рррр")

	validateDate := func(entry *widget.Entry, target *time.Time) {
		raw := entry.Text
		if raw == "" {
			*target = time.Time{}
			entry.SetValidationError(nil)
			return
		}
		t, err := time.Parse(dateLayout, raw)
		if err != nil {
			entry.SetValidationError(errors.New("формат: дд.мм.рррр"))
			return
		}
		*target = t
		entry.SetValidationError(nil)
	}

	setRange := func(months int) {
		now := time.Now()
		from := now.AddDate(0, -months, 0)
		fromEntry.SetText(from.Format(dateLayout))
		toEntry.SetText(now.Format(dateLayout))
	}

	rangeButtons := container.NewHBox(
		widget.NewButton("1 міс", func() { setRange(1) }),
		widget.NewButton("3 міс", func() { setRange(3) }),
		widget.NewButton("6 міс", func() { setRange(6) }),
		widget.NewButton("Рік", func() { setRange(12) }),
		widget.NewButton("Весь час", func() {
			fromEntry.SetText("")
			toEntry.SetText("")
		}),
	)

	// --- Секції звіту ---
	includeSummary := widget.NewCheck("Підсумок", func(v bool) { opts.includeSummary = v })
	includeCategory := widget.NewCheck("За категоріями", func(v bool) { opts.includeCategory = v })
	includeMonthly := widget.NewCheck("По місяцях", func(v bool) { opts.includeMonthly = v })
	includeChart := widget.NewCheck("Графік", func(v bool) { opts.includeChart = v })

	includeSummary.SetChecked(true)
	includeCategory.SetChecked(true)
	includeMonthly.SetChecked(true)
	includeChart.SetChecked(true)

	// --- Попередній перегляд ---
	previewLabel := widget.NewRichTextFromMarkdown("")

	refreshPreview := func() {
		txs := filterByDate(state.GetTransactions(), opts.from, opts.to)
		if len(txs) == 0 {
			previewLabel.ParseMarkdown("_Немає транзакцій для обраного діапазону_")
			return
		}
		income, expense := calcTotals(txs)
		previewLabel.ParseMarkdown(fmt.Sprintf(
			"**Транзакцій:** %d\n\n**Надходження:** %s грн\n\n**Витрати:** %s грн\n\n**Баланс:** %s грн",
			len(txs),
			income.StringFixed(2),
			expense.StringFixed(2),
			income.Sub(expense).StringFixed(2),
		))
	}

	fromEntry.OnChanged = func(s string) {
		validateDate(fromEntry, &opts.from)
		refreshPreview()
	}
	toEntry.OnChanged = func(s string) {
		validateDate(toEntry, &opts.to)
		refreshPreview()
	}

	refreshPreview()

	// --- Прогрес ---
	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// runInBackground виконує fn в горутині, блокує кнопки і показує прогрес.
	runInBackground := func(buttons []*widget.Button, fn func() error, onDone func(err error)) {
		for _, btn := range buttons {
			btn.Disable()
		}
		fyne.Do(func() {
			progress.Show()
			statusLabel.SetText("Генерація...")
		})
		go func() {
			err := fn()
			fyne.Do(func() {
				progress.Hide()
				for _, btn := range buttons {
					btn.Enable()
				}
				onDone(err)
			})
		}()
	}

	// --- Кнопка повного звіту ---
	var generateBtn *widget.Button
	var exportDTOBtn *widget.Button

	generateBtn = widget.NewButtonWithIcon("Згенерувати звіт", theme.DocumentSaveIcon(), func() {
		if len(state.GetTransactions()) == 0 {
			dialog.ShowInformation("Немає даних", "Спочатку імпортуйте банківські виписки.", win)
			return
		}

		d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if writer == nil {
				return
			}
			outputPath := writer.URI().Path()
			writer.Close()

			report := buildReport(state, opts)

			runInBackground(
				[]*widget.Button{generateBtn, exportDTOBtn},
				func() error {
					return generateReport(formatSelect.Selected, opts.templatePath, report, outputPath)
				},
				func(err error) {
					if err != nil {
						statusLabel.SetText("Помилка генерації")
						dialog.ShowError(err, win)
						return
					}
					statusLabel.SetText(fmt.Sprintf("Збережено: %s", outputPath))
					dialog.ShowInformation("Готово", "Звіт успішно збережено!", win)
				},
			)
		}, win)

		d.SetFileName(defaultFileName(formatSelect.Selected))
		d.SetFilter(newFormatFilter(formatSelect.Selected))
		d.Show()
	})
	generateBtn.Importance = widget.HighImportance

	// --- Кнопка експорту через DTO ---
	exportDTOBtn = widget.NewButtonWithIcon("Експорт (DTO)", theme.DownloadIcon(), func() {
		txs := state.GetTransactions()
		if len(txs) == 0 {
			dialog.ShowInformation("Немає даних", "Спочатку імпортуйте банківські виписки.", win)
			return
		}

		// Конвертуємо у DTO до відкриття діалогу збереження —
		// щоб не тримати посилання на стан під час асинхронної операції.
		filtered := filterByDate(txs, opts.from, opts.to)
		dtos := models.ToTransactions(filtered)

		d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if writer == nil {
				return
			}
			outputPath := writer.URI().Path()
			writer.Close()

			selectedFormat := formatSelect.Selected
			runInBackground(
				[]*widget.Button{generateBtn, exportDTOBtn},
				func() error {
					return generateReportDTO(selectedFormat, opts.templatePath, dtos, outputPath)
				},
				func(genErr error) {
					if genErr != nil {
						statusLabel.SetText("Помилка експорту")
						dialog.ShowError(genErr, win)
						return
					}
					statusLabel.SetText(fmt.Sprintf("DTO експорт збережено: %s", outputPath))
					dialog.ShowInformation("Готово",
						fmt.Sprintf("Експортовано %d транзакцій.", len(dtos)), win)
				},
			)
		}, win)

		ext := "xlsx"
		if formatSelect.Selected == "ODS" {
			ext = "ods"
		}
		d.SetFileName("export_dto_" + time.Now().Format("2006-01-02") + "." + ext)
		d.SetFilter(newFormatFilter(formatSelect.Selected))
		d.Show()
	})
	exportDTOBtn.Importance = widget.MediumImportance

	formatSelect.OnChanged = func(s string) {
		opts.templatePath = ""
		templateLabel.SetText("не обрано")
		templateLabel.Importance = widget.LowImportance
		templateLabel.Refresh()
	}

	// --- Layout ---
	settingsCard := widget.NewCard("Налаштування", "", container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Формат:"),
			formatSelect,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Шаблон:"),
			container.NewHBox(templateLabel, templateBtn, clearTemplateBtn),
		),
		widget.NewSeparator(),
		widget.NewLabel("Період:"),
		container.NewGridWithColumns(3,
			fromEntry,
			widget.NewLabelWithStyle("—", fyne.TextAlignCenter, fyne.TextStyle{}),
			toEntry,
		),
		rangeButtons,
	))

	sectionsCard := widget.NewCard("Секції звіту", "", container.NewGridWithColumns(2,
		includeSummary,
		includeCategory,
		includeMonthly,
		includeChart,
	))

	previewCard := widget.NewCard("Попередній перегляд", "", previewLabel)

	// Пояснення різниці між кнопками
	dtoHint := widget.NewLabel("Експорт (DTO) — спрощений формат: ID, дата, тип, сума, валюта, контрагент, категорія.")
	dtoHint.Importance = widget.LowImportance
	dtoHint.Wrapping = fyne.TextWrapWord

	bottom := container.NewVBox(
		widget.NewSeparator(),
		progress,
		statusLabel,
		container.NewGridWithColumns(2, generateBtn, exportDTOBtn),
		dtoHint,
	)

	content := container.NewVScroll(container.NewVBox(
		settingsCard,
		sectionsCard,
		previewCard,
	))

	return container.NewBorder(nil, bottom, nil, nil, content)
}

// --- Бізнес-логіка ---

func buildReport(state *AppState, opts reportOptions) *models.Report {
	txs := filterByDate(state.GetTransactions(), opts.from, opts.to)

	report := &models.Report{
		Transactions: txs,
		ByCategory:   make(map[string]models.CategorySummary),
		ByMonth:      make(map[string]models.MonthSummary),
	}

	if len(txs) == 0 {
		return report
	}

	report.Period.From = txs[0].Date
	report.Period.To = txs[0].Date

	var totalExpense decimal.Decimal

	for _, tx := range txs {
		if tx.Date.Before(report.Period.From) {
			report.Period.From = tx.Date
		}
		if tx.Date.After(report.Period.To) {
			report.Period.To = tx.Date
		}

		if tx.Type == models.Credit {
			report.TotalIncome = report.TotalIncome.Add(tx.Amount)
		} else {
			report.TotalExpense = report.TotalExpense.Add(tx.Amount)
			totalExpense = totalExpense.Add(tx.Amount)
		}

		if opts.includeCategory {
			cs := report.ByCategory[tx.Category]
			cs.Category = tx.Category
			cs.Count++
			cs.Total = cs.Total.Add(tx.Amount)
			report.ByCategory[tx.Category] = cs
		}

		if opts.includeMonthly {
			key := tx.Date.Format("2006-01")
			ms := report.ByMonth[key]
			ms.Month = key
			if tx.Type == models.Credit {
				ms.Income = ms.Income.Add(tx.Amount)
			} else {
				ms.Expense = ms.Expense.Add(tx.Amount)
			}
			report.ByMonth[key] = ms
		}
	}

	report.NetBalance = report.TotalIncome.Sub(report.TotalExpense)

	if opts.includeCategory && !totalExpense.IsZero() {
		for key, cs := range report.ByCategory {
			f, _ := cs.Total.Div(totalExpense).Mul(decimal.NewFromInt(100)).Float64()
			cs.Percent = f
			report.ByCategory[key] = cs
		}
	}

	return report
}

func generateReport(format, templatePath string, report *models.Report, outputPath string) error {
	switch format {
	case "XLSX":
		r := &reports.XLSXReporter{TemplatePath: templatePath}
		return r.Generate(report, outputPath)
	case "ODS":
		r := &reports.ODSReporter{}
		return r.Generate(report, outputPath)
	default:
		return fmt.Errorf("невідомий формат: %s", format)
	}
}

func generateReportDTO(format, templatePath string, dtos []models.TransactionDTO, outputPath string) error {
	switch format {
	case "XLSX":
		r := &reports.XLSXReporter{TemplatePath: templatePath}
		return r.GenerateFromDTO(dtos, outputPath)
	case "ODS":
		r := &reports.ODSReporter{}
		return r.GenerateFromDTO(dtos, outputPath)
	default:
		return fmt.Errorf("невідомий формат: %s", format)
	}
}

func filterByDate(txs []*models.Transaction, from, to time.Time) []*models.Transaction {
	if from.IsZero() && to.IsZero() {
		return txs
	}
	result := make([]*models.Transaction, 0, len(txs))
	for _, tx := range txs {
		if !from.IsZero() && tx.Date.Before(from) {
			continue
		}
		if !to.IsZero() && tx.Date.After(to) {
			continue
		}
		result = append(result, tx)
	}
	return result
}

func calcTotals(txs []*models.Transaction) (income, expense decimal.Decimal) {
	for _, tx := range txs {
		if tx.Type == models.Credit {
			income = income.Add(tx.Amount)
		} else {
			expense = expense.Add(tx.Amount)
		}
	}
	return
}

func defaultFileName(format string) string {
	date := time.Now().Format("2006-01-02")
	if format == "ODS" {
		return fmt.Sprintf("звіт_%s.ods", date)
	}
	return fmt.Sprintf("звіт_%s.xlsx", date)
}

func newFormatFilter(format string) storage.FileFilter {
	if format == "ODS" {
		return storage.NewExtensionFileFilter([]string{".ods"})
	}
	return storage.NewExtensionFileFilter([]string{".xlsx"})
}
