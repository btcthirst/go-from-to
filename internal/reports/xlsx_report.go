package reports

import (
	"bank-analyzer/internal/models"
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

const (
	colorHeader   = "4472C4"
	colorIncome   = "E2EFDA"
	colorExpense  = "FCE4D6"
	colorSubtotal = "D9E1F2"
)

var (
	ErrOpenTemplate = errors.New("failed to open template excel file")
	ErrWriteData    = errors.New("failed to write data to excel")
	ErrSaveReport   = errors.New("failed to save report")
)

type XLSXReporter struct {
	TemplatePath string
}

// Generate генерує повний звіт з моделі Report (з усіма листами та графіком).
func (r *XLSXReporter) Generate(report *models.Report, outputPath string) (err error) {
	f, err := r.openOrCreate()
	if err != nil {
		return err
	}
	defer func() {
		if errClose := f.Close(); err == nil {
			err = errClose
		}
	}()

	if err := r.writeTransactions(f, report.Transactions); err != nil {
		return fmt.Errorf("%w(transactions): %v", ErrWriteData, err)
	}
	if err := r.writeSummary(f, report); err != nil {
		return fmt.Errorf("%w(summary): %v", ErrWriteData, err)
	}
	if err := r.writePivot(f, report); err != nil {
		return fmt.Errorf("%w(pivot): %v", ErrWriteData, err)
	}
	if err := r.writeMonthlyChart(f, report); err != nil {
		return fmt.Errorf("%w(monthly chart): %v", ErrWriteData, err)
	}

	return f.SaveAs(outputPath)
}

// GenerateFromDTO генерує спрощений звіт з []TransactionDTO.
// Створює один лист «Транзакції» без зведень та графіків.
// Корисно для швидкого експорту без побудови повного models.Report.
func (r *XLSXReporter) GenerateFromDTO(dtos []models.TransactionDTO, outputPath string) (err error) {
	f, err := r.openOrCreate()
	if err != nil {
		return err
	}
	defer func() {
		if errClose := f.Close(); err == nil {
			err = errClose
		}
	}()

	if err := r.writeDTOs(f, dtos); err != nil {
		return fmt.Errorf("%w(dto): %v", ErrWriteData, err)
	}

	return f.SaveAs(outputPath)
}

// openOrCreate відкриває шаблон або створює новий файл.
func (r *XLSXReporter) openOrCreate() (*excelize.File, error) {
	if r.TemplatePath != "" {
		f, err := excelize.OpenFile(r.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOpenTemplate, err)
		}
		return f, nil
	}
	return excelize.NewFile(), nil
}

// writeDTOs записує []TransactionDTO на лист «Транзакції».
// Колонки: ID, Дата, Тип, Сума, Валюта, Контрагент, Категорія.
func (r *XLSXReporter) writeDTOs(f *excelize.File, dtos []models.TransactionDTO) error {
	const sheet = "Транзакції"

	// Видаляємо дефолтний Sheet1 якщо він є і ми створюємо новий файл
	if idx, err := f.GetSheetIndex("Sheet1"); err == nil && idx >= 0 {
		f.DeleteSheet("Sheet1")
	}

	f.NewSheet(sheet)
	f.SetActiveSheet(mustSheetIndex(f, sheet))

	// --- Стилі ---
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{colorHeader}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Family: "Arial", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: false},
		Border:    thinBorder(),
	})

	incomeStyle, _ := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{colorIncome}, Pattern: 1},
		Font:   &excelize.Font{Family: "Arial", Size: 10},
		Border: thinBorder(),
	})

	expenseStyle, _ := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{colorExpense}, Pattern: 1},
		Font:   &excelize.Font{Family: "Arial", Size: 10},
		Border: thinBorder(),
	})

	/*amountStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font:      &excelize.Font{Family: "Arial", Size: 10},
		NumFmt:    4, // #,##0.00
		Alignment: &excelize.Alignment{Horizontal: "right"},
		Border:    thinBorder(),
	})*/

	incomeAmtStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{colorIncome}, Pattern: 1},
		Font:      &excelize.Font{Family: "Arial", Size: 10},
		NumFmt:    4,
		Alignment: &excelize.Alignment{Horizontal: "right"},
		Border:    thinBorder(),
	})

	expenseAmtStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{colorExpense}, Pattern: 1},
		Font:      &excelize.Font{Family: "Arial", Size: 10},
		NumFmt:    4,
		Alignment: &excelize.Alignment{Horizontal: "right"},
		Border:    thinBorder(),
	})

	// --- Заголовки ---
	type col struct {
		header string
		width  float64
	}
	cols := []col{
		{"ID", 12},
		{"Дата", 12},
		{"Тип", 14},
		{"Сума", 14},
		{"Валюта", 9},
		{"Контрагент", 30},
		{"Категорія", 20},
	}

	for i, c := range cols {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, c.header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, colName, colName, c.width)
	}
	f.SetRowHeight(sheet, 1, 18)

	// --- Дані ---
	for rowIdx, dto := range dtos {
		row := rowIdx + 2

		isIncome := dto.Type == models.Credit
		rowStyle := expenseStyle
		if isIncome {
			rowStyle = incomeStyle
		}
		amtRowStyle := expenseAmtStyle
		if isIncome {
			amtRowStyle = incomeAmtStyle
		}

		values := []interface{}{
			dto.ID,
			dto.Date,
			dto.TypeLabel(),
			dto.Amount,
			dto.Currency,
			dto.Counterparty,
			dto.Category,
		}

		for colIdx, val := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			f.SetCellValue(sheet, cell, val)

			// Колонка «Сума» (індекс 3) — числовий стиль з форматом
			if colIdx == 3 {
				f.SetCellStyle(sheet, cell, cell, amtRowStyle)
			} else {
				f.SetCellStyle(sheet, cell, cell, rowStyle)
			}
		}

		f.SetRowHeight(sheet, row, 16)
	}

	// --- Автофільтр та закріплення ---
	lastRow := len(dtos) + 1
	f.AutoFilter(
		sheet,
		fmt.Sprintf("A1:G%d", lastRow),
		[]excelize.AutoFilterOptions{},
	)
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// --- Підсумковий рядок ---
	if len(dtos) > 0 {
		summaryRow := lastRow + 1
		f.SetRowHeight(sheet, summaryRow, 18)

		subtotalStyle, _ := f.NewStyle(&excelize.Style{
			Fill:      excelize.Fill{Type: "pattern", Color: []string{colorSubtotal}, Pattern: 1},
			Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 10},
			NumFmt:    4,
			Alignment: &excelize.Alignment{Horizontal: "right"},
			Border:    thinBorder(),
		})
		subtotalLabelStyle, _ := f.NewStyle(&excelize.Style{
			Fill:   excelize.Fill{Type: "pattern", Color: []string{colorSubtotal}, Pattern: 1},
			Font:   &excelize.Font{Bold: true, Family: "Arial", Size: 10},
			Border: thinBorder(),
		})

		// «Разом» у колонці C
		totalCell, _ := excelize.CoordinatesToCellName(3, summaryRow)
		f.SetCellValue(sheet, totalCell, "Разом:")
		f.SetCellStyle(sheet, totalCell, totalCell, subtotalLabelStyle)

		// Формула суми в колонці D
		sumCell, _ := excelize.CoordinatesToCellName(4, summaryRow)
		f.SetCellFormula(sheet, sumCell, fmt.Sprintf("=SUBTOTAL(9,D2:D%d)", lastRow))
		f.SetCellStyle(sheet, sumCell, sumCell, subtotalStyle)

		// Кількість у колонці A
		countCell, _ := excelize.CoordinatesToCellName(1, summaryRow)
		f.SetCellFormula(sheet, countCell, fmt.Sprintf("=COUNTA(A2:A%d)&\" записів\"", lastRow))
		f.SetCellStyle(sheet, countCell, countCell, subtotalLabelStyle)
	}

	return nil
}

// --- Існуючі методи без змін ---

func (r *XLSXReporter) writeTransactions(f *excelize.File, transactions []*models.Transaction) error {
	sheet := "Транзакції"
	f.NewSheet(sheet)

	headers := []string{"Дата", "Тип", "Сума", "Валюта", "Категорія", "Опис", "Контрагент", "Баланс"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{colorHeader}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	customFmt := `#,##0.00`
	incomeStyle, _ := f.NewStyle(&excelize.Style{
		Fill:         excelize.Fill{Type: "pattern", Color: []string{colorIncome}, Pattern: 1},
		CustomNumFmt: &customFmt,
	})
	expenseStyle, _ := f.NewStyle(&excelize.Style{
		Fill:         excelize.Fill{Type: "pattern", Color: []string{colorExpense}, Pattern: 1},
		CustomNumFmt: &customFmt,
	})

	for i, tx := range transactions {
		row := i + 2
		typeStr := "Надходження"
		style := incomeStyle
		if tx.Type == models.Debit {
			typeStr = "Витрата"
			style = expenseStyle
		}

		data := []interface{}{
			tx.Date.Format("02.01.2006"),
			typeStr,
			tx.Amount.InexactFloat64(),
			tx.Currency,
			tx.Category,
			tx.Description,
			tx.Counterparty,
			tx.Balance.InexactFloat64(),
		}
		for col, val := range data {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, val)
		}
		f.SetRowStyle(sheet, row, row, style)
	}

	f.AutoFilter(
		sheet,
		fmt.Sprintf("A1:H%d", len(transactions)+1),
		[]excelize.AutoFilterOptions{},
	)
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	return nil
}

func (r *XLSXReporter) writeSummary(f *excelize.File, report *models.Report) error {
	sheet := "Підсумок"
	f.NewSheet(sheet)

	data := [][]interface{}{
		{"Показник", "Значення"},
		{"Надходження", report.TotalIncome.InexactFloat64()},
		{"Витрати", report.TotalExpense.InexactFloat64()},
		{"Баланс", report.NetBalance.InexactFloat64()},
		{"Транзакцій", len(report.Transactions)},
		{"Період з", report.Period.From.Format("02.01.2006")},
		{"Період по", report.Period.To.Format("02.01.2006")},
	}
	for i, row := range data {
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue(sheet, cell, val)
		}
	}
	return nil
}

func (r *XLSXReporter) writePivot(f *excelize.File, report *models.Report) error {
	sheet := "За категоріями"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "Категорія")
	f.SetCellValue(sheet, "B1", "Сума")
	f.SetCellValue(sheet, "C1", "% від витрат")
	f.SetCellValue(sheet, "D1", "Кількість")

	row := 2
	for _, summary := range report.ByCategory {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), summary.Category)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), summary.Total.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", summary.Percent))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), summary.Count)
		row++
	}
	return nil
}

func (r *XLSXReporter) writeMonthlyChart(f *excelize.File, report *models.Report) error {
	sheet := "Графік"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "Місяць")
	f.SetCellValue(sheet, "B1", "Надходження")
	f.SetCellValue(sheet, "C1", "Витрати")

	row := 2
	for month, ms := range report.ByMonth {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), month)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), ms.Income.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), ms.Expense.InexactFloat64())
		row++
	}

	chart := &excelize.Chart{
		Type: excelize.Bar,
		Series: []excelize.ChartSeries{
			{
				Name:       sheet + "!$B$1",
				Categories: fmt.Sprintf("%s!$A$2:$A$%d", sheet, row-1),
				Values:     fmt.Sprintf("%s!$B$2:$B$%d", sheet, row-1),
			},
			{
				Name:       sheet + "!$C$1",
				Categories: fmt.Sprintf("%s!$A$2:$A$%d", sheet, row-1),
				Values:     fmt.Sprintf("%s!$C$2:$C$%d", sheet, row-1),
			},
		},
		Title: []excelize.RichTextRun{{Text: "Рух коштів по місяцях"}},
	}
	return f.AddChart(sheet, "E2", chart)
}

// --- Допоміжні функції ---

func mustSheetIndex(f *excelize.File, sheet string) int {
	idx, err := f.GetSheetIndex(sheet)
	if err != nil || idx < 0 {
		return 0
	}
	return idx
}

func thinBorder() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "CCCCCC", Style: 1},
		{Type: "right", Color: "CCCCCC", Style: 1},
		{Type: "top", Color: "CCCCCC", Style: 1},
		{Type: "bottom", Color: "CCCCCC", Style: 1},
	}
}
