package reports

// Звіт "Журнал-ордер 311" у форматі XLSX.
//
// Заголовки фіксовані як на зразку:
//   A  №п/п        B  Постачальник   C  Дата
//   D  Дт рах.311/Сума              E  Оборот по Дт  (=D{row})
//   F…  рахунки з Report311Config
//   ?   Оборот по Кт                               (=SUM(F:last))
//
// Конфігурація передається ззовні через mappings.Report311Config
// (завантажується в config.Load з assets/report_311.yaml).
//
// Залежності: github.com/xuri/excelize/v2

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/xuri/excelize/v2"

	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"
)

// ─── Репортер ─────────────────────────────────────────────────────────────────

type XLSX311Reporter struct {
	cfg mappings.Report311Config
}

// NewXLSX311Reporter створює репортер з конфігурацією зі звичайного місця:
//
//	cfg := config.Load()
//	r   := reports.NewXLSX311Reporter(cfg.Report311)
func NewXLSX311Reporter(cfg mappings.Report311Config) *XLSX311Reporter {
	return &XLSX311Reporter{cfg: cfg}
}

// Generate формує звіт за місяць із зрізу транзакцій.
func (r *XLSX311Reporter) Generate(txs []*models.Transaction, outputPath string) error {
	// Сортування по даті ASC
	sorted := make([]*models.Transaction, len(txs))
	copy(sorted, txs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	// Допоміжні мапи з конфігу
	mainSet := make(map[string]bool, len(r.cfg.MainCategories))
	for _, c := range r.cfg.MainCategories {
		mainSet[c] = true
	}
	subIdx := make(map[string]int, len(r.cfg.SubColumns))
	for i, sc := range r.cfg.SubColumns {
		subIdx[sc.Category] = i
	}

	// Номери колонок (1-based):
	// A=1 B=2 C=3 D=4 E=5 F=6 ... (5+subCount) totCol
	subCount := len(r.cfg.SubColumns)
	totCol := 5 + subCount + 1

	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Журнал-ордер 311"
	f.SetSheetName("Sheet1", sheet)

	st := build311Styles(f)
	write311Headers(f, sheet, r.cfg, subCount, totCol, st)
	lastData := write311Data(f, sheet, sorted, mainSet, subIdx, subCount, totCol, st)
	write311Totals(f, sheet, lastData+1, lastData, totCol, st)
	apply311Settings(f, sheet, lastData+1, totCol)

	return f.SaveAs(outputPath)
}

// ─── Стилі ───────────────────────────────────────────────────────────────────

type s311 struct {
	header, body, bodyAlt, bodyNum, bodyNumAlt, total, totalNum int
}

func build311Styles(f *excelize.File) s311 {
	mk := func(s *excelize.Style) int { id, _ := f.NewStyle(s); return id }
	nf := p311(`#,##0.00;-#,##0.00;"-"`)

	border := []excelize.Border{
		{Type: "left", Color: "AAAAAA", Style: 1},
		{Type: "right", Color: "AAAAAA", Style: 1},
		{Type: "top", Color: "AAAAAA", Style: 1},
		{Type: "bottom", Color: "AAAAAA", Style: 1},
	}
	return s311{
		header: mk(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 9, Color: "FFFFFF", Family: "Arial"},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"4472C4"}},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border:    border,
		}),
		body: mk(&excelize.Style{
			Font:      &excelize.Font{Size: 9, Family: "Arial"},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
			Border:    border,
		}),
		bodyAlt: mk(&excelize.Style{
			Font:      &excelize.Font{Size: 9, Family: "Arial"},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
			Border:    border,
		}),
		bodyNum: mk(&excelize.Style{
			Font:         &excelize.Font{Size: 9, Family: "Arial"},
			Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
			Border:       border,
			CustomNumFmt: nf,
		}),
		bodyNumAlt: mk(&excelize.Style{
			Font:         &excelize.Font{Size: 9, Family: "Arial"},
			Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
			Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
			Border:       border,
			CustomNumFmt: nf,
		}),
		total: mk(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 9, Color: "FFFFFF", Family: "Arial"},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"4472C4"}},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border:    border,
		}),
		totalNum: mk(&excelize.Style{
			Font:         &excelize.Font{Bold: true, Size: 9, Family: "Arial"},
			Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DCE6F1"}},
			Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
			Border:       border,
			CustomNumFmt: nf,
		}),
	}
}

// ─── Заголовки ────────────────────────────────────────────────────────────────

func write311Headers(f *excelize.File, sheet string, cfg mappings.Report311Config, subCount, totCol int, st s311) {
	// Об'єднання рядків 1–2 для фіксованих колонок
	for _, m := range [][2]string{
		{"A1", "A2"}, {"B1", "B2"}, {"C1", "C2"},
		{"D1", "D2"}, {"E1", "E2"},
		{a311(totCol, 1), a311(totCol, 2)},
	} {
		f.MergeCell(sheet, m[0], m[1])
	}
	// Об'єднання рядка 1 для групи sub-колонок
	if subCount > 0 {
		f.MergeCell(sheet, a311(6, 1), a311(5+subCount, 1))
	}

	// Фіксовані заголовки рядка 1
	for cell, val := range map[string]string{
		"A1":            "№\nп/п",
		"B1":            "Постачальник",
		"C1":            "Дата",
		"D1":            "Дт рах.311\nСума",
		"E1":            "Оборот\nпо Дт",
		a311(6, 1):      "С Кт 311 в Дт рахунків",
		a311(totCol, 1): "Оборот\nпо Кт",
	} {
		f.SetCellValue(sheet, cell, val)
		f.SetCellStyle(sheet, cell, cell, st.header)
	}

	// Рядок 2: назви рахунків з конфігу (F2, G2, ...)
	for i, sc := range cfg.SubColumns {
		addr := a311(6+i, 2)
		f.SetCellValue(sheet, addr, sc.Account)
		f.SetCellStyle(sheet, addr, addr, st.header)
	}

	// Висота рядків
	f.SetRowHeight(sheet, 1, 28)
	f.SetRowHeight(sheet, 2, 18)

	// Ширини колонок
	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "B", "B", 22)
	f.SetColWidth(sheet, "C", "C", 11)
	f.SetColWidth(sheet, "D", "D", 13)
	f.SetColWidth(sheet, "E", "E", 11)
	for i := 0; i < subCount; i++ {
		col, _ := excelize.ColumnNumberToName(6 + i)
		f.SetColWidth(sheet, col, col, 9)
	}
	totColName, _ := excelize.ColumnNumberToName(totCol)
	f.SetColWidth(sheet, totColName, totColName, 13)
}

// ─── Дані ─────────────────────────────────────────────────────────────────────

func write311Data(
	f *excelize.File,
	sheet string,
	txs []*models.Transaction,
	mainSet map[string]bool,
	subIdx map[string]int,
	subCount, totCol int,
	st s311,
) int {
	const dataStart = 3
	row := dataStart

	for i, tx := range txs {
		alt := i%2 == 1
		txtSt := st.body
		numSt := st.bodyNum
		if alt {
			txtSt = st.bodyAlt
			numSt = st.bodyNumAlt
		}

		amt, _ := tx.Amount.Float64()

		setv311(f, sheet, 1, row, i+1, txtSt)
		setv311(f, sheet, 2, row, tx.Counterparty, txtSt)
		setv311(f, sheet, 3, row, tx.Date.Format("02.01.2006"), txtSt)

		dAddr := a311(4, row)
		eAddr := a311(5, row)

		switch {
		case mainSet[tx.Category]:
			f.SetCellValue(sheet, dAddr, amt)
			f.SetCellStyle(sheet, dAddr, dAddr, numSt)
			for ci := 0; ci < subCount; ci++ {
				a := a311(6+ci, row)
				f.SetCellValue(sheet, a, nil)
				f.SetCellStyle(sheet, a, a, numSt)
			}

		case func() bool { _, ok := subIdx[tx.Category]; return ok }():
			idx := subIdx[tx.Category]
			f.SetCellValue(sheet, dAddr, nil)
			f.SetCellStyle(sheet, dAddr, dAddr, numSt)
			for ci := 0; ci < subCount; ci++ {
				a := a311(6+ci, row)
				if ci == idx {
					f.SetCellValue(sheet, a, amt)
				} else {
					f.SetCellValue(sheet, a, nil)
				}
				f.SetCellStyle(sheet, a, a, numSt)
			}

		default:
			f.SetCellValue(sheet, dAddr, nil)
			f.SetCellStyle(sheet, dAddr, dAddr, numSt)
			for ci := 0; ci < subCount; ci++ {
				a := a311(6+ci, row)
				f.SetCellValue(sheet, a, nil)
				f.SetCellStyle(sheet, a, a, numSt)
			}
		}

		// E = =D{row}
		f.SetCellFormula(sheet, eAddr, dAddr)
		f.SetCellStyle(sheet, eAddr, eAddr, numSt)

		// Оборот по Кт = =SUM(F{row}:lastSub{row})
		fName, _ := excelize.ColumnNumberToName(6)
		lastSubName, _ := excelize.ColumnNumberToName(5 + subCount)
		f.SetCellFormula(sheet, a311(totCol, row),
			fmt.Sprintf("SUM(%s%d:%s%d)", fName, row, lastSubName, row))
		f.SetCellStyle(sheet, a311(totCol, row), a311(totCol, row), numSt)

		row++
	}
	return row - 1
}

// ─── Підсумки ─────────────────────────────────────────────────────────────────

func write311Totals(f *excelize.File, sheet string, totalRow, lastData, totCol int, st s311) {
	f.MergeCell(sheet, a311(1, totalRow), a311(3, totalRow))
	f.SetCellValue(sheet, a311(1, totalRow), "РАЗОМ")
	f.SetCellStyle(sheet, a311(1, totalRow), a311(3, totalRow), st.total)

	for col := 4; col <= totCol; col++ {
		letter, _ := excelize.ColumnNumberToName(col)
		addr := a311(col, totalRow)
		f.SetCellFormula(sheet, addr,
			fmt.Sprintf("SUM(%s3:%s%d)", letter, letter, lastData))
		f.SetCellStyle(sheet, addr, addr, st.totalNum)
	}
}

// ─── Налаштування ─────────────────────────────────────────────────────────────

func apply311Settings(f *excelize.File, sheet string, totalRow, totCol int) {
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      2,
		TopLeftCell: "A3",
		ActivePane:  "bottomLeft",
	})
	lastCol, _ := excelize.ColumnNumberToName(totCol)
	f.AutoFilter(sheet, fmt.Sprintf("A2:%s%d", lastCol, totalRow), nil)
}

// ─── Утиліти ─────────────────────────────────────────────────────────────────

func a311(col, row int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return name + strconv.Itoa(row)
}

func setv311(f *excelize.File, sheet string, col, row int, val interface{}, styleID int) {
	addr := a311(col, row)
	f.SetCellValue(sheet, addr, val)
	f.SetCellStyle(sheet, addr, addr, styleID)
}

func p311(s string) *string { return &s }
