package reports

// ODS генератор. Всі XML-константи — прості рядки без змінних всередині,
// namespace скопійовані побайтово з реального файлу LibreOffice 26.2.

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"bank-analyzer/internal/models"
)

const odsMimeType = "application/vnd.oasis.opendocument.spreadsheet"

// odsStylesXML — повний styles.xml, згенерований Python і перевірений як валідний XML.
// Namespace скопійовано з реального файлу LibreOffice 26.2, версія 1.4.
const odsStylesXML = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><office:document-styles xmlns:presentation=\"urn:oasis:names:tc:opendocument:xmlns:presentation:1.0\" xmlns:css3t=\"http://www.w3.org/TR/css3-text/\" xmlns:grddl=\"http://www.w3.org/2003/g/data-view#\" xmlns:xhtml=\"http://www.w3.org/1999/xhtml\" xmlns:dom=\"http://www.w3.org/2001/xml-events\" xmlns:script=\"urn:oasis:names:tc:opendocument:xmlns:script:1.0\" xmlns:form=\"urn:oasis:names:tc:opendocument:xmlns:form:1.0\" xmlns:math=\"http://www.w3.org/1998/Math/MathML\" xmlns:office=\"urn:oasis:names:tc:opendocument:xmlns:office:1.0\" xmlns:ooo=\"http://openoffice.org/2004/office\" xmlns:fo=\"urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0\" xmlns:ooow=\"http://openoffice.org/2004/writer\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" xmlns:drawooo=\"http://openoffice.org/2010/draw\" xmlns:oooc=\"http://openoffice.org/2004/calc\" xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:calcext=\"urn:org:documentfoundation:names:experimental:calc:xmlns:calcext:1.0\" xmlns:style=\"urn:oasis:names:tc:opendocument:xmlns:style:1.0\" xmlns:text=\"urn:oasis:names:tc:opendocument:xmlns:text:1.0\" xmlns:of=\"urn:oasis:names:tc:opendocument:xmlns:of:1.2\" xmlns:tableooo=\"http://openoffice.org/2009/table\" xmlns:draw=\"urn:oasis:names:tc:opendocument:xmlns:drawing:1.0\" xmlns:dr3d=\"urn:oasis:names:tc:opendocument:xmlns:dr3d:1.0\" xmlns:rpt=\"http://openoffice.org/2005/report\" xmlns:svg=\"urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0\" xmlns:chart=\"urn:oasis:names:tc:opendocument:xmlns:chart:1.0\" xmlns:table=\"urn:oasis:names:tc:opendocument:xmlns:table:1.0\" xmlns:meta=\"urn:oasis:names:tc:opendocument:xmlns:meta:1.0\" xmlns:loext=\"urn:org:documentfoundation:names:experimental:office:xmlns:loext:1.0\" xmlns:number=\"urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0\" xmlns:field=\"urn:openoffice:names:experimental:ooo-ms-interop:xmlns:field:1.0\" office:version=\"1.4\"><office:font-face-decls><style:font-face style:name=\"Calibri\" svg:font-family=\"Calibri\" style:font-family-generic=\"swiss\"/></office:font-face-decls><office:styles><style:default-style style:family=\"table-cell\"><style:text-properties style:font-name=\"Calibri\" fo:font-size=\"11pt\" fo:language=\"uk\" fo:country=\"UA\"/></style:default-style><style:style style:name=\"Default\" style:family=\"table-cell\"/></office:styles><office:automatic-styles><number:number-style style:name=\"N2\"><number:number number:decimal-places=\"2\" number:min-decimal-places=\"2\" number:min-integer-digits=\"1\" number:grouping=\"true\"/></number:number-style></office:automatic-styles><office:master-styles><style:master-page style:name=\"Default\" style:page-layout-name=\"Mpm1\"/></office:master-styles></office:document-styles>"

// odsContentOpen — відкриваючий тег content.xml з повним набором namespace.
const odsContentOpen = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><office:document-content xmlns:presentation=\"urn:oasis:names:tc:opendocument:xmlns:presentation:1.0\" xmlns:css3t=\"http://www.w3.org/TR/css3-text/\" xmlns:grddl=\"http://www.w3.org/2003/g/data-view#\" xmlns:xhtml=\"http://www.w3.org/1999/xhtml\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns:xforms=\"http://www.w3.org/2002/xforms\" xmlns:dom=\"http://www.w3.org/2001/xml-events\" xmlns:script=\"urn:oasis:names:tc:opendocument:xmlns:script:1.0\" xmlns:form=\"urn:oasis:names:tc:opendocument:xmlns:form:1.0\" xmlns:math=\"http://www.w3.org/1998/Math/MathML\" xmlns:office=\"urn:oasis:names:tc:opendocument:xmlns:office:1.0\" xmlns:ooo=\"http://openoffice.org/2004/office\" xmlns:fo=\"urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0\" xmlns:ooow=\"http://openoffice.org/2004/writer\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" xmlns:drawooo=\"http://openoffice.org/2010/draw\" xmlns:oooc=\"http://openoffice.org/2004/calc\" xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:calcext=\"urn:org:documentfoundation:names:experimental:calc:xmlns:calcext:1.0\" xmlns:style=\"urn:oasis:names:tc:opendocument:xmlns:style:1.0\" xmlns:text=\"urn:oasis:names:tc:opendocument:xmlns:text:1.0\" xmlns:of=\"urn:oasis:names:tc:opendocument:xmlns:of:1.2\" xmlns:tableooo=\"http://openoffice.org/2009/table\" xmlns:draw=\"urn:oasis:names:tc:opendocument:xmlns:drawing:1.0\" xmlns:dr3d=\"urn:oasis:names:tc:opendocument:xmlns:dr3d:1.0\" xmlns:rpt=\"http://openoffice.org/2005/report\" xmlns:formx=\"urn:openoffice:names:experimental:ooxml-odf-interop:xmlns:form:1.0\" xmlns:svg=\"urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0\" xmlns:chart=\"urn:oasis:names:tc:opendocument:xmlns:chart:1.0\" xmlns:table=\"urn:oasis:names:tc:opendocument:xmlns:table:1.0\" xmlns:meta=\"urn:oasis:names:tc:opendocument:xmlns:meta:1.0\" xmlns:loext=\"urn:org:documentfoundation:names:experimental:office:xmlns:loext:1.0\" xmlns:number=\"urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0\" xmlns:field=\"urn:openoffice:names:experimental:ooo-ms-interop:xmlns:field:1.0\" office:version=\"1.4\">"

const odsManifest = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
	"<manifest:manifest xmlns:manifest=\"urn:oasis:names:tc:opendocument:xmlns:manifest:1.0\" manifest:version=\"1.4\" xmlns:loext=\"urn:org:documentfoundation:names:experimental:office:xmlns:loext:1.0\">\n" +
	"<manifest:file-entry manifest:full-path=\"\" manifest:version=\"1.4\" manifest:media-type=\"application/vnd.oasis.opendocument.spreadsheet\"/>\n" +
	"<manifest:file-entry manifest:full-path=\"Configurations2/\" manifest:media-type=\"application/vnd.sun.xml.ui.configuration\"/>\n" +
	"<manifest:file-entry manifest:full-path=\"styles.xml\" manifest:media-type=\"text/xml\"/>\n" +
	"<manifest:file-entry manifest:full-path=\"manifest.rdf\" manifest:media-type=\"application/rdf+xml\"/>\n" +
	"<manifest:file-entry manifest:full-path=\"content.xml\" manifest:media-type=\"text/xml\"/>\n" +
	"<manifest:file-entry manifest:full-path=\"meta.xml\" manifest:media-type=\"text/xml\"/>\n" +
	"<manifest:file-entry manifest:full-path=\"settings.xml\" manifest:media-type=\"text/xml\"/>\n" +
	"</manifest:manifest>"

const odsManifestRDF = "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n" +
	"<rdf:RDF xmlns:rdf=\"http://www.w3.org/1999/02/22-rdf-syntax-ns#\">\n" +
	"<rdf:Description rdf:about=\"styles.xml\"><rdf:type rdf:resource=\"http://docs.oasis-open.org/ns/office/1.2/meta/odf#StylesFile\"/></rdf:Description>\n" +
	"<rdf:Description rdf:about=\"\"><ns0:hasPart xmlns:ns0=\"http://docs.oasis-open.org/ns/office/1.2/meta/pkg#\" rdf:resource=\"styles.xml\"/></rdf:Description>\n" +
	"<rdf:Description rdf:about=\"content.xml\"><rdf:type rdf:resource=\"http://docs.oasis-open.org/ns/office/1.2/meta/odf#ContentFile\"/></rdf:Description>\n" +
	"<rdf:Description rdf:about=\"\"><ns0:hasPart xmlns:ns0=\"http://docs.oasis-open.org/ns/office/1.2/meta/pkg#\" rdf:resource=\"content.xml\"/></rdf:Description>\n" +
	"<rdf:Description rdf:about=\"\"><rdf:type rdf:resource=\"http://docs.oasis-open.org/ns/office/1.2/meta/pkg#Document\"/></rdf:Description>\n" +
	"</rdf:RDF>"

const odsSettings = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" +
	"<office:document-settings xmlns:office=\"urn:oasis:names:tc:opendocument:xmlns:office:1.0\" " +
	"xmlns:config=\"urn:oasis:names:tc:opendocument:xmlns:config:1.0\" " +
	"xmlns:ooo=\"http://openoffice.org/2004/office\" office:version=\"1.4\">" +
	"<office:settings>" +
	"<config:config-item-set config:name=\"ooo:view-settings\">" +
	"<config:config-item config:name=\"ActiveTable\" config:type=\"string\">Транзакції</config:config-item>" +
	"</config:config-item-set>" +
	"</office:settings>" +
	"</office:document-settings>"

// ─── Публічний API ────────────────────────────────────────────────────────────

type ODSReporter struct{}

func (r *ODSReporter) Generate(report *models.Report, outputPath string) error {
	b := newODSBuilder()
	b.addSheet("Транзакції", buildTransactionRows(report.Transactions))
	b.addSheet("Підсумок", buildSummaryRows(report))
	b.addSheet("За категоріями", buildCategoryRows(report))
	b.addSheet("По місяцях", buildMonthlyRows(report))
	return b.writeTo(outputPath)
}

func (r *ODSReporter) GenerateFromDTO(dtos []models.TransactionDTO, outputPath string) error {
	b := newODSBuilder()
	b.addSheet("Транзакції", buildDTORows(dtos))
	return b.writeTo(outputPath)
}

// ─── Рядки таблиць ───────────────────────────────────────────────────────────

func buildTransactionRows(txs []*models.Transaction) []odsRow {
	rows := make([]odsRow, 0, len(txs)+1)
	rows = append(rows, headerRow("Дата", "Тип", "Сума", "Валюта", "Категорія", "Опис", "Контрагент", "Баланс"))
	for _, tx := range txs {
		isDebit := tx.Type == models.Debit
		typeStr := "Надходження"
		if isDebit {
			typeStr = "Витрата"
		}
		amt, _ := tx.Amount.Float64()
		bal, _ := tx.Balance.Float64()
		rows = append(rows, odsRow{
			debit: isDebit,
			cells: []odsCell{
				textCell(tx.Date.Format("02.01.2006")),
				textCell(typeStr),
				numCell(amt),
				textCell(tx.Currency),
				textCell(tx.Category),
				textCell(tx.Description),
				textCell(tx.Counterparty),
				numCell(bal),
			},
		})
	}
	return rows
}

func buildDTORows(dtos []models.TransactionDTO) []odsRow {
	rows := make([]odsRow, 0, len(dtos)+1)
	rows = append(rows, headerRow("ID", "Дата", "Тип", "Сума", "Валюта", "Контрагент", "Категорія"))
	for _, dto := range dtos {
		isDebit := dto.Type == models.Debit
		rows = append(rows, odsRow{
			debit: isDebit,
			cells: []odsCell{
				textCell(dto.ID),
				textCell(dto.Date),
				textCell(dto.TypeLabel()),
				numCell(dto.Amount),
				textCell(dto.Currency),
				textCell(dto.Counterparty),
				textCell(dto.Category),
			},
		})
	}
	return rows
}

func buildSummaryRows(report *models.Report) []odsRow {
	inc, _ := report.TotalIncome.Float64()
	exp, _ := report.TotalExpense.Float64()
	net, _ := report.NetBalance.Float64()
	return []odsRow{
		headerRow("Показник", "Значення"),
		{cells: []odsCell{textCell("Надходження"), numCell(inc)}},
		{cells: []odsCell{textCell("Витрати"), numCell(exp)}},
		{cells: []odsCell{textCell("Баланс"), numCell(net)}},
		{cells: []odsCell{textCell("Транзакцій"), textCell(fmt.Sprintf("%d", len(report.Transactions)))}},
		{cells: []odsCell{textCell("Період з"), textCell(report.Period.From.Format("02.01.2006"))}},
		{cells: []odsCell{textCell("Період по"), textCell(report.Period.To.Format("02.01.2006"))}},
	}
}

func buildCategoryRows(report *models.Report) []odsRow {
	rows := []odsRow{headerRow("Категорія", "Сума", "% від витрат", "Кількість")}
	for _, cs := range report.ByCategory {
		total, _ := cs.Total.Float64()
		rows = append(rows, odsRow{cells: []odsCell{
			textCell(cs.Category), numCell(total), numCell(cs.Percent), numCell(float64(cs.Count)),
		}})
	}
	return rows
}

func buildMonthlyRows(report *models.Report) []odsRow {
	rows := []odsRow{headerRow("Місяць", "Надходження", "Витрати")}
	for _, ms := range report.ByMonth {
		inc, _ := ms.Income.Float64()
		exp, _ := ms.Expense.Float64()
		rows = append(rows, odsRow{cells: []odsCell{
			textCell(ms.Month), numCell(inc), numCell(exp),
		}})
	}
	return rows
}

// ─── Типи ────────────────────────────────────────────────────────────────────

type odsCell struct {
	text   string
	isNum  bool
	number float64
}

func textCell(s string) odsCell { return odsCell{text: s} }
func numCell(v float64) odsCell { return odsCell{isNum: true, number: v} }

type odsRow struct {
	cells  []odsCell
	header bool
	debit  bool
}

func headerRow(cols ...string) odsRow {
	cells := make([]odsCell, len(cols))
	for i, c := range cols {
		cells[i] = textCell(c)
	}
	return odsRow{header: true, cells: cells}
}

// ─── Builder ─────────────────────────────────────────────────────────────────

type odsSheet struct {
	name string
	rows []odsRow
}

type odsBuilder struct{ sheets []odsSheet }

func newODSBuilder() *odsBuilder { return &odsBuilder{} }

func (b *odsBuilder) addSheet(name string, rows []odsRow) {
	b.sheets = append(b.sheets, odsSheet{name: name, rows: rows})
}

func (b *odsBuilder) writeTo(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("не вдалося створити файл: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// 1. mimetype — ПЕРШИЙ, ZIP_STORED
	mh := &zip.FileHeader{Name: "mimetype", Method: zip.Store, Modified: time.Unix(0, 0)}
	mh.Extra = nil

	mw, err := zw.CreateHeader(mh)
	if err != nil {
		return err
	}
	mw.Write([]byte(odsMimeType))

	// 2. Configurations2/ — порожня директорія
	zw.CreateHeader(&zip.FileHeader{Name: "Configurations2/", Method: zip.Store})

	// 3–7. Інші файли (порядок відповідає реальному LibreOffice файлу)
	if err := zs(zw, "styles.xml", odsStylesXML); err != nil {
		return err
	}
	if err := zs(zw, "manifest.rdf", odsManifestRDF); err != nil {
		return err
	}
	if err := zs(zw, "content.xml", b.buildContent()); err != nil {
		return err
	}
	if err := zs(zw, "meta.xml", b.buildMeta()); err != nil {
		return err
	}
	if err := zs(zw, "settings.xml", odsSettings); err != nil {
		return err
	}
	return zs(zw, "META-INF/manifest.xml", odsManifest)
}

func zs(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, content)
	return err
}

// ─── Генерація XML ───────────────────────────────────────────────────────────

func (b *odsBuilder) buildMeta() string {
	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" +
		"<office:document-meta xmlns:grddl=\"http://www.w3.org/2003/g/data-view#\"" +
		" xmlns:meta=\"urn:oasis:names:tc:opendocument:xmlns:meta:1.0\"" +
		" xmlns:dc=\"http://purl.org/dc/elements/1.1/\"" +
		" xmlns:xlink=\"http://www.w3.org/1999/xlink\"" +
		" xmlns:ooo=\"http://openoffice.org/2004/office\"" +
		" xmlns:office=\"urn:oasis:names:tc:opendocument:xmlns:office:1.0\"" +
		" office:version=\"1.4\">" +
		"<office:meta>" +
		"<meta:generator>bank-analyzer</meta:generator>" +
		"<dc:language>uk-UA</dc:language>" +
		"<dc:date>" + time.Now().Format(time.RFC3339) + "</dc:date>" +
		"</office:meta>" +
		"</office:document-meta>"
}

func (b *odsBuilder) buildContent() string {
	var sb strings.Builder
	sb.WriteString(odsContentOpen)
	sb.WriteString("<office:automatic-styles>")
	sb.WriteString(odsCellStyles)
	sb.WriteString("</office:automatic-styles>")
	sb.WriteString("<office:body><office:spreadsheet>")
	for _, sheet := range b.sheets {
		sb.WriteString("<table:table table:name=\"")
		sb.WriteString(xmlEsc(sheet.name))
		sb.WriteString("\">")
		for _, row := range sheet.rows {
			sb.WriteString(renderODSRow(row))
		}
		sb.WriteString("</table:table>")
	}
	sb.WriteString("</office:spreadsheet></office:body>")
	sb.WriteString("</office:document-content>")
	return sb.String()
}

// odsCellStyles — стилі клітинок в automatic-styles content.xml.
const odsCellStyles = "" +
	"<style:style style:name=\"ceHeader\" style:family=\"table-cell\" style:parent-style-name=\"Default\">" +
	"<style:table-cell-properties fo:background-color=\"#4472C4\" style:vertical-align=\"middle\"/>" +
	"<style:text-properties fo:color=\"#FFFFFF\" fo:font-weight=\"bold\" fo:font-size=\"10pt\" style:font-name=\"Calibri\"/>" +
	"</style:style>" +
	"<style:style style:name=\"ceIncome\" style:family=\"table-cell\" style:parent-style-name=\"Default\">" +
	"<style:table-cell-properties fo:background-color=\"#E2EFDA\"/>" +
	"<style:text-properties fo:font-size=\"10pt\" style:font-name=\"Calibri\"/>" +
	"</style:style>" +
	"<style:style style:name=\"ceExpense\" style:family=\"table-cell\" style:parent-style-name=\"Default\">" +
	"<style:table-cell-properties fo:background-color=\"#FCE4D6\"/>" +
	"<style:text-properties fo:font-size=\"10pt\" style:font-name=\"Calibri\"/>" +
	"</style:style>" +
	"<style:style style:name=\"ceIncomeNum\" style:family=\"table-cell\" style:parent-style-name=\"Default\" style:data-style-name=\"N2\">" +
	"<style:table-cell-properties fo:background-color=\"#E2EFDA\"/>" +
	"<style:text-properties fo:font-size=\"10pt\" style:font-name=\"Calibri\"/>" +
	"</style:style>" +
	"<style:style style:name=\"ceExpenseNum\" style:family=\"table-cell\" style:parent-style-name=\"Default\" style:data-style-name=\"N2\">" +
	"<style:table-cell-properties fo:background-color=\"#FCE4D6\"/>" +
	"<style:text-properties fo:font-size=\"10pt\" style:font-name=\"Calibri\"/>" +
	"</style:style>"

func renderODSRow(row odsRow) string {
	var sb strings.Builder
	sb.WriteString("<table:table-row>")
	for _, cell := range row.cells {
		sb.WriteString(renderODSCell(cell, row.header, row.debit))
	}
	sb.WriteString("</table:table-row>")
	return sb.String()
}

func renderODSCell(cell odsCell, header, debit bool) string {
	var style string
	switch {
	case header:
		style = "ceHeader"
	case cell.isNum && debit:
		style = "ceExpenseNum"
	case cell.isNum:
		style = "ceIncomeNum"
	case debit:
		style = "ceExpense"
	default:
		style = "ceIncome"
	}
	if !cell.isNum {
		return "<table:table-cell office:value-type=\"string\" table:style-name=\"" + style + "\">" +
			"<text:p>" + xmlEsc(cell.text) + "</text:p>" +
			"</table:table-cell>"
	}
	val := fmt.Sprintf("%g", cell.number)
	return "<table:table-cell office:value-type=\"float\" office:value=\"" + val + "\" table:style-name=\"" + style + "\">" +
		"<text:p>" + val + "</text:p>" +
		"</table:table-cell>"
}

func xmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
