package reports

// ODS генератор — підхід "шаблон + патч content.xml".
//
// Алгоритм:
//   1. Відкрити template.ods як ZIP (go:embed)
//   2. Збудувати новий content.xml зі своїми даними
//   3. Переписати ZIP: всі файли копіюються as-is, content.xml — замінюється
//
// Чому так, а не xml.Encoder:
//   Go encoder генерує <body xmlns="office:..."> замість <office:body> —
//   LibreOffice не читає такий XML. Тому content.xml будується через
//   strings.Builder з явними тегами, а xml.EscapeText — тільки для значень.
//
// Чому шаблон, а не генерація з нуля:
//   ODF вимагає щоб mimetype був першим файлом у ZIP, записаним як ZIP_STORED
//   з CRC у локальному заголовку (не в Data Descriptor). LibreOffice правильно
//   створює такий архів. Якщо генерувати ZIP самостійно через zip.NewWriter —
//   Go ставить прапор 0x0008 і залишає нулі в заголовку, що призводить до
//   попередження "пошкоджений файл" при відкритті.

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"bank-analyzer/internal/models"
)

//go:embed template.ods
var odsTemplate []byte

// ─── Публічний API ────────────────────────────────────────────────────────────

type ODSReporter struct{}

func (r *ODSReporter) Generate(report *models.Report, outputPath string) error {
	sheets := []odsSheet{
		{name: "Транзакції", rows: buildTransactionRows(report.Transactions)},
		{name: "Підсумок", rows: buildSummaryRows(report)},
		{name: "За категоріями", rows: buildCategoryRows(report)},
		{name: "По місяцях", rows: buildMonthlyRows(report)},
	}
	return patchAndWrite(sheets, outputPath)
}

func (r *ODSReporter) GenerateFromDTO(dtos []models.TransactionDTO, outputPath string) error {
	sheets := []odsSheet{
		{name: "Транзакції", rows: buildDTORows(dtos)},
	}
	return patchAndWrite(sheets, outputPath)
}

// ─── ZIP патч ────────────────────────────────────────────────────────────────

// patchAndWrite читає шаблон, замінює content.xml, записує результат.
func patchAndWrite(sheets []odsSheet, outputPath string) error {
	zr, err := zip.NewReader(bytes.NewReader(odsTemplate), int64(len(odsTemplate)))
	if err != nil {
		return fmt.Errorf("не вдалося відкрити шаблон: %w", err)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("не вдалося створити файл: %w", err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	newContent := []byte(buildContentXML(sheets, time.Now()))

	for _, f := range zr.File {
		fh := f.FileHeader // копіюємо заголовок (метод, прапори, дати)

		w, err := zw.CreateHeader(&fh)
		if err != nil {
			return fmt.Errorf("zip CreateHeader %s: %w", f.Name, err)
		}

		if f.Name == "content.xml" {
			if _, err = w.Write(newContent); err != nil {
				return fmt.Errorf("zip write content.xml: %w", err)
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("zip open %s: %w", f.Name, err)
		}
		_, copyErr := io.Copy(w, rc)
		rc.Close()
		if copyErr != nil {
			return fmt.Errorf("zip copy %s: %w", f.Name, copyErr)
		}
	}

	return nil
}

// ─── Побудова content.xml ─────────────────────────────────────────────────────

type odsSheet struct {
	name string
	rows []odsRow
}

func buildContentXML(sheets []odsSheet, _ time.Time) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<office:document-content`)
	sb.WriteString(` xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"`)
	sb.WriteString(` xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"`)
	sb.WriteString(` xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"`)
	sb.WriteString(` xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"`)
	sb.WriteString(` xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"`)
	sb.WriteString(` xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0"`)
	sb.WriteString(` xmlns:calcext="urn:org:documentfoundation:names:experimental:calc:xmlns:calcext:1.0"`)
	sb.WriteString(` office:version="1.4">`)

	// automatic-styles порожній — всі стилі визначені у шаблоні (styles.xml)
	sb.WriteString(`<office:automatic-styles/>`)

	sb.WriteString(`<office:body><office:spreadsheet>`)
	for _, sheet := range sheets {
		writeSheet(&sb, sheet)
	}
	sb.WriteString(`</office:spreadsheet></office:body>`)
	sb.WriteString(`</office:document-content>`)

	return sb.String()
}

func writeSheet(sb *strings.Builder, sheet odsSheet) {
	sb.WriteString(`<table:table table:name="`)
	sb.WriteString(xmlAttr(sheet.name))
	sb.WriteString(`">`)

	for _, row := range sheet.rows {
		sb.WriteString(`<table:table-row>`)
		for _, cell := range row.cells {
			writeCell(sb, cell, row.style(cell))
		}
		sb.WriteString(`</table:table-row>`)
	}

	sb.WriteString(`</table:table>`)
}

func writeCell(sb *strings.Builder, cell odsCell, styleName string) {
	if cell.isNum {
		sb.WriteString(`<table:table-cell office:value-type="float"`)
		sb.WriteString(` calcext:value-type="float"`)
		sb.WriteString(` office:value="`)
		sb.WriteString(cell.rawVal) // без наукової нотації, повна точність
		sb.WriteString(`" table:style-name="`)
		sb.WriteString(styleName)
		sb.WriteString(`"><text:p>`)
		sb.WriteString(cell.dispVal) // 2 знаки після коми для відображення
		sb.WriteString(`</text:p></table:table-cell>`)
	} else {
		sb.WriteString(`<table:table-cell office:value-type="string"`)
		sb.WriteString(` table:style-name="`)
		sb.WriteString(styleName)
		sb.WriteString(`"><text:p>`)
		sb.WriteString(xmlText(cell.text))
		sb.WriteString(`</text:p></table:table-cell>`)
	}
}

// ─── Типи рядків і клітинок ───────────────────────────────────────────────────

type odsCell struct {
	text    string
	rawVal  string // office:value — повна точність, формат 'f'
	dispVal string // text:p       — 2 знаки після коми
	isNum   bool
	isDebit bool // використовується лише в headerRow для ігнорування
}

type odsRow struct {
	cells  []odsCell
	header bool
	debit  bool
}

// style повертає назву стилю клітинки зі шаблону.
// Назви мають збігатися зі стилями у template.ods.
func (r odsRow) style(c odsCell) string {
	switch {
	case r.header:
		return "ceHeader"
	case r.debit && c.isNum:
		return "ceExpenseNum"
	case r.debit:
		return "ceExpense"
	case c.isNum:
		return "ceIncomeNum"
	default:
		return "ceIncome"
	}
}

func strCell(s string) odsCell { return odsCell{text: s} }

func floatCell(v float64) odsCell {
	return odsCell{
		isNum:   true,
		rawVal:  strconv.FormatFloat(v, 'f', -1, 64),
		dispVal: strconv.FormatFloat(v, 'f', 2, 64),
	}
}

func hdrRow(cols ...string) odsRow {
	cells := make([]odsCell, len(cols))
	for i, c := range cols {
		cells[i] = strCell(c)
	}
	return odsRow{header: true, cells: cells}
}

func dataRow(debit bool, cells ...odsCell) odsRow {
	return odsRow{debit: debit, cells: cells}
}

// ─── Рядки таблиць ───────────────────────────────────────────────────────────

func buildTransactionRows(txs []*models.Transaction) []odsRow {
	rows := make([]odsRow, 0, len(txs)+1)
	rows = append(rows, hdrRow("Дата", "Тип", "Сума", "Валюта", "Категорія", "Опис", "Контрагент", "Баланс"))
	for _, tx := range txs {
		isDebit := tx.Type == models.Debit
		typeStr := "Надходження"
		if isDebit {
			typeStr = "Витрата"
		}
		amt, _ := tx.Amount.Float64()
		bal, _ := tx.Balance.Float64()
		rows = append(rows, dataRow(isDebit,
			strCell(tx.Date.Format("02.01.2006")),
			strCell(typeStr),
			floatCell(amt),
			strCell(tx.Currency),
			strCell(tx.Category),
			strCell(tx.Description),
			strCell(tx.Counterparty),
			floatCell(bal),
		))
	}
	return rows
}

func buildDTORows(dtos []models.TransactionDTO) []odsRow {
	rows := make([]odsRow, 0, len(dtos)+1)
	rows = append(rows, hdrRow("ID", "Дата", "Тип", "Сума", "Валюта", "Контрагент", "Категорія"))
	for _, dto := range dtos {
		isDebit := dto.Type == models.Debit
		rows = append(rows, dataRow(isDebit,
			strCell(dto.ID),
			strCell(dto.Date),
			strCell(dto.TypeLabel()),
			floatCell(dto.Amount),
			strCell(dto.Currency),
			strCell(dto.Counterparty),
			strCell(dto.Category),
		))
	}
	return rows
}

func buildSummaryRows(report *models.Report) []odsRow {
	inc, _ := report.TotalIncome.Float64()
	exp, _ := report.TotalExpense.Float64()
	net, _ := report.NetBalance.Float64()
	return []odsRow{
		hdrRow("Показник", "Значення"),
		dataRow(false, strCell("Надходження"), floatCell(inc)),
		dataRow(false, strCell("Витрати"), floatCell(exp)),
		dataRow(false, strCell("Баланс"), floatCell(net)),
		dataRow(false, strCell("Транзакцій"), strCell(strconv.Itoa(len(report.Transactions)))),
		dataRow(false, strCell("Період з"), strCell(report.Period.From.Format("02.01.2006"))),
		dataRow(false, strCell("Період по"), strCell(report.Period.To.Format("02.01.2006"))),
	}
}

func buildCategoryRows(report *models.Report) []odsRow {
	keys := make([]string, 0, len(report.ByCategory))
	for k := range report.ByCategory {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([]odsRow, 0, len(keys)+1)
	rows = append(rows, hdrRow("Категорія", "Сума", "% від витрат", "Кількість"))
	for _, k := range keys {
		cs := report.ByCategory[k]
		total, _ := cs.Total.Float64()
		rows = append(rows, dataRow(false,
			strCell(cs.Category),
			floatCell(total),
			floatCell(cs.Percent),
			floatCell(float64(cs.Count)),
		))
	}
	return rows
}

func buildMonthlyRows(report *models.Report) []odsRow {
	keys := make([]string, 0, len(report.ByMonth))
	for k := range report.ByMonth {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([]odsRow, 0, len(keys)+1)
	rows = append(rows, hdrRow("Місяць", "Надходження", "Витрати"))
	for _, k := range keys {
		ms := report.ByMonth[k]
		inc, _ := ms.Income.Float64()
		exp, _ := ms.Expense.Float64()
		rows = append(rows, dataRow(false,
			strCell(ms.Month),
			floatCell(inc),
			floatCell(exp),
		))
	}
	return rows
}

// ─── XML хелпери ─────────────────────────────────────────────────────────────

// xmlText екранує текст клітинки (&, <, >, тощо).
func xmlText(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// xmlAttr екранує значення атрибута (додатково лапки).
func xmlAttr(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return strings.ReplaceAll(buf.String(), `"`, "&quot;")
}
