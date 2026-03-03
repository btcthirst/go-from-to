package ui

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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

	// ─── Загальний звіт ──────────────────────────────────────────────────────

	opts := reportOptions{
		includeSummary:  true,
		includeCategory: true,
		includeMonthly:  true,
		includeChart:    true,
	}

	formatSelect := widget.NewSelect([]string{"XLSX", "ODS"}, nil)
	formatSelect.SetSelected("XLSX")

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

	includeSummary := widget.NewCheck("Підсумок", func(v bool) { opts.includeSummary = v })
	includeCategory := widget.NewCheck("За категоріями", func(v bool) { opts.includeCategory = v })
	includeMonthly := widget.NewCheck("По місяцях", func(v bool) { opts.includeMonthly = v })
	includeChart := widget.NewCheck("Графік", func(v bool) { opts.includeChart = v })

	includeSummary.SetChecked(true)
	includeCategory.SetChecked(true)
	includeMonthly.SetChecked(true)
	includeChart.SetChecked(true)

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

	fromEntry.OnChanged = func(s string) { validateDate(fromEntry, &opts.from); refreshPreview() }
	toEntry.OnChanged = func(s string) { validateDate(toEntry, &opts.to); refreshPreview() }
	refreshPreview()

	formatSelect.OnChanged = func(s string) {
		opts.templatePath = ""
		templateLabel.SetText("не обрано")
		templateLabel.Importance = widget.LowImportance
		templateLabel.Refresh()
	}

	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

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

	exportDTOBtn = widget.NewButtonWithIcon("Експорт (DTO)", theme.DownloadIcon(), func() {
		txs := state.GetTransactions()
		if len(txs) == 0 {
			dialog.ShowInformation("Немає даних", "Спочатку імпортуйте банківські виписки.", win)
			return
		}
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

	dtoHint := widget.NewLabel("Експорт (DTO) — спрощений формат: ID, дата, тип, сума, валюта, контрагент, категорія.")
	dtoHint.Importance = widget.LowImportance
	dtoHint.Wrapping = fyne.TextWrapWord

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
		includeSummary, includeCategory, includeMonthly, includeChart,
	))

	previewCard := widget.NewCard("Попередній перегляд", "", previewLabel)

	generalBottom := container.NewVBox(
		container.NewGridWithColumns(2, generateBtn, exportDTOBtn),
		dtoHint,
	)

	generalSection := container.NewVBox(settingsCard, sectionsCard, previewCard, generalBottom)

	// ─── Звіт 311 ────────────────────────────────────────────────────────────

	report311Section := newReport311Section(state, win, runInBackground)

	// ─── Tabs ─────────────────────────────────────────────────────────────────

	tabs := container.NewAppTabs(
		container.NewTabItem("Журнал-ордер 311", container.NewVScroll(report311Section)),
		container.NewTabItem("Загальний звіт", container.NewVScroll(generalSection)),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	bottom := container.NewVBox(
		widget.NewSeparator(),
		progress,
		statusLabel,
	)

	return container.NewBorder(nil, bottom, nil, nil, tabs)
}

// newReport311Section будує UI-секцію для звіту "Журнал-ордер 311".
func newReport311Section(
	state *AppState,
	win fyne.Window,
	runInBackground func([]*widget.Button, func() error, func(error)),
) fyne.CanvasObject {

	// --- Вибір місяця ---
	now := time.Now()

	monthEntry := widget.NewEntry()
	monthEntry.SetText(now.Format("01.2006"))
	monthEntry.SetPlaceHolder("мм.рррр")

	var selectedMonth time.Time
	parseMonth := func() (time.Time, error) {
		return time.Parse("01.2006", monthEntry.Text)
	}

	// Кнопки швидкого вибору місяця
	prevMonthBtn := widget.NewButton("← Попередній", func() {
		t, err := parseMonth()
		if err != nil {
			t = now
		}
		monthEntry.SetText(t.AddDate(0, -1, 0).Format("01.2006"))
	})
	curMonthBtn := widget.NewButton("Поточний", func() {
		monthEntry.SetText(now.Format("01.2006"))
	})
	nextMonthBtn := widget.NewButton("Наступний →", func() {
		t, err := parseMonth()
		if err != nil {
			t = now
		}
		monthEntry.SetText(t.AddDate(0, 1, 0).Format("01.2006"))
	})

	// --- Попередній перегляд ---
	previewLabel := widget.NewLabel("")
	previewLabel.Importance = widget.LowImportance

	refreshPreview311 := func() {
		t, err := parseMonth()
		if err != nil {
			previewLabel.SetText("Невірний формат місяця (мм.рррр)")
			return
		}
		selectedMonth = t

		all := state.GetTransactions()
		txs := filterByMonth(all, selectedMonth)

		if len(txs) == 0 {
			if len(all) == 0 {
				previewLabel.SetText("Транзакції не імпортовані")
			} else {
				previewLabel.SetText(fmt.Sprintf(
					"Транзакцій за %s не знайдено.\nДоступні місяці: %s",
					selectedMonth.Format("01.2006"),
					uniqueMonths(all),
				))
			}
			return
		}
		income, expense := calcTotals(txs)
		previewLabel.SetText(fmt.Sprintf(
			"Транзакцій: %d   Надходження: %s грн   Витрати: %s грн",
			len(txs), income.StringFixed(2), expense.StringFixed(2),
		))
	}

	monthEntry.OnChanged = func(_ string) { refreshPreview311() }
	refreshPreview311()

	// Оновлюємо preview коли з'являються нові транзакції
	state.AddTransactionListener(func() {
		refreshPreview311()
	})

	// --- Кнопка генерації ---
	var generate311Btn *widget.Button

	generate311Btn = widget.NewButtonWithIcon("Згенерувати Журнал-ордер 311", theme.DocumentSaveIcon(), func() {
		t, err := parseMonth()
		if err != nil {
			dialog.ShowError(errors.New("невірний формат місяця, очікується мм.рррр"), win)
			return
		}
		selectedMonth = t

		txs := filterByMonth(state.GetTransactions(), selectedMonth)
		if len(txs) == 0 {
			dialog.ShowInformation("Немає даних",
				fmt.Sprintf("Транзакцій за %s не знайдено.", selectedMonth.Format("01.2006")), win)
			return
		}

		d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, saveErr error) {
			if saveErr != nil {
				dialog.ShowError(saveErr, win)
				return
			}
			if writer == nil {
				return
			}
			outputPath := writer.URI().Path()
			writer.Close()

			reporter := reports.NewXLSX311Reporter(state.Config.Report311)
			snapshot := make([]*models.Transaction, len(txs))
			copy(snapshot, txs)

			runInBackground(
				[]*widget.Button{generate311Btn},
				func() error {
					return reporter.Generate(snapshot, outputPath)
				},
				func(genErr error) {
					if genErr != nil {
						dialog.ShowError(genErr, win)
						return
					}
					dialog.ShowInformation("Готово",
						fmt.Sprintf("Журнал-ордер 311 за %s збережено.\nТранзакцій: %d",
							selectedMonth.Format("січень 2006"), len(snapshot)), win)
				},
			)
		}, win)

		fileName := fmt.Sprintf("311_%s.xlsx", selectedMonth.Format("01_2006"))
		d.SetFileName(fileName)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
		d.Show()
	})
	generate311Btn.Importance = widget.HighImportance

	// --- Інформація про конфігурацію ---
	cfg := state.Config.Report311
	mainCats := fmt.Sprintf("Основні категорії (Дт 311): %v", cfg.MainCategories)
	subCols := ""
	for _, sc := range cfg.SubColumns {
		subCols += fmt.Sprintf("  рах.%s → %s\n", sc.Account, sc.Category)
	}

	configInfo := widget.NewLabel(mainCats + "\n\nРозбивка по рахунках:\n" + subCols)
	configInfo.Importance = widget.LowImportance
	configInfo.Wrapping = fyne.TextWrapWord

	configCard := widget.NewCard("Конфігурація звіту", "з assets/report_311.yaml", configInfo)

	monthCard := widget.NewCard("Період", "", container.NewVBox(
		container.NewGridWithColumns(3, prevMonthBtn, curMonthBtn, nextMonthBtn),
		container.NewGridWithColumns(2,
			widget.NewLabel("Місяць (мм.рррр):"),
			monthEntry,
		),
		previewLabel,
	))

	return container.NewVBox(
		monthCard,
		configCard,
		generate311Btn,
	)
}

// --- Бізнес-логіка ---

func buildReport(state *AppState, opts reportOptions) *models.Report {
	txs := filterByDate(state.GetTransactions(), opts.from, opts.to)
	return reports.BuildReport(txs, reports.BuildOptions{
		IncludeCategory: opts.includeCategory,
		IncludeMonthly:  opts.includeMonthly,
	})
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

// filterByMonth повертає транзакції за конкретний місяць (рік + місяць).
func filterByMonth(txs []*models.Transaction, month time.Time) []*models.Transaction {
	result := make([]*models.Transaction, 0)
	for _, tx := range txs {
		if tx.Date.Year() == month.Year() && tx.Date.Month() == month.Month() {
			result = append(result, tx)
		}
	}
	return result
}

// uniqueMonths повертає відсортований рядок унікальних місяців з транзакцій.
// Використовується для діагностики коли фільтр повертає порожній результат.
func uniqueMonths(txs []*models.Transaction) string {
	seen := make(map[string]struct{})
	for _, tx := range txs {
		seen[tx.Date.Format("01.2006")] = struct{}{}
	}
	months := make([]string, 0, len(seen))
	for m := range seen {
		months = append(months, m)
	}
	sort.Strings(months)
	return strings.Join(months, ", ")
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
