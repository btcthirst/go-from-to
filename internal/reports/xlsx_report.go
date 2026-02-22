package reports

import (
	"bank-analyzer/internal/models"
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// Colors for Excel cells styling ???
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

func (r *XLSXReporter) Generate(report *models.Report, outputPath string) (err error) {
	var f *excelize.File
	if r.TemplatePath != "" {

		f, err = excelize.OpenFile(r.TemplatePath)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrOpenTemplate, err)
		}
	} else {
		f = excelize.NewFile()
	}
	defer func() {
		errClose := f.Close()
		if err == nil {
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

func (r *XLSXReporter) writeTransactions(f *excelize.File, transactions []*models.Transaction) error {
	sheet := "Транзакції"
	f.NewSheet(sheet)

	// Заголовки
	headers := []string{"Дата", "Тип", "Сума", "Валюта", "Категорія", "Опис", "Контрагент", "Баланс"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Стиль заголовку
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{colorHeader}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	incomeStyle, _ := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{colorIncome}, Pattern: 1},
		NumFmt: 177, // #,##0.00
	})
	expenseStyle, _ := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{colorExpense}, Pattern: 1},
		NumFmt: 177,
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

	// Автофільтр
	f.AutoFilter(
		sheet,
		fmt.Sprintf("A1:H%d", len(transactions)+1),
		[]excelize.AutoFilterOptions{},
	)
	// Заморозити перший рядок
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
	// Дані для графіку
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

	// Стовпчиковий графік
	chart := &excelize.Chart{
		Type: excelize.Bar,
		Series: []excelize.ChartSeries{
			{Name: sheet + "!$B$1", Categories: fmt.Sprintf("%s!$A$2:$A$%d", sheet, row-1),
				Values: fmt.Sprintf("%s!$B$2:$B$%d", sheet, row-1)},
			{Name: sheet + "!$C$1", Categories: fmt.Sprintf("%s!$A$2:$A$%d", sheet, row-1),
				Values: fmt.Sprintf("%s!$C$2:$C$%d", sheet, row-1)},
		},
		Title: []excelize.RichTextRun{{Text: "Рух коштів по місяцях"}},
	}
	return f.AddChart(sheet, "E2", chart)
}
