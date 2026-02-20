package writer

import (
	"archive/zip"
	"bytes"
	"excel-parser/internal/model"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// WriteResultsODS записує результати у файл ODS
func WriteResultsODS(results []model.ResultRecord, outputFile string) error {
	// Створюємо буфер для ZIP архіву
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	defer zw.Close()

	// Записуємо mimetype
	mimeType, err := zw.Create("mimetype")
	if err != nil {
		return fmt.Errorf("помилка при створенні mimetype: %w", err)
	}
	mimeType.Write([]byte("application/vnd.oasis.opendocument.spreadsheet"))

	// Записуємо META-INF/manifest.xml
	manifest, err := zw.Create("META-INF/manifest.xml")
	if err != nil {
		return fmt.Errorf("помилка при створенні manifest: %w", err)
	}
	manifest.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
 <manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/>
 <manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/>
 <manifest:file-entry manifest:full-path="meta.xml" manifest:media-type="text/xml"/>
 <manifest:file-entry manifest:full-path="styles.xml" manifest:media-type="text/xml"/>
 <manifest:file-entry manifest:full-path="settings.xml" manifest:media-type="text/xml"/>
</manifest:manifest>`))

	// Записуємо meta.xml
	meta, err := zw.Create("meta.xml")
	if err != nil {
		return fmt.Errorf("помилка при створенні meta: %w", err)
	}
	meta.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0" office:version="1.2">
 <office:meta>
  <meta:creation-date>` + time.Now().Format("2006-01-02T15:04:05") + `</meta:creation-date>
 </office:meta>
</office:document-meta>`))

	// Записуємо styles.xml
	styles, err := zw.Create("styles.xml")
	if err != nil {
		return fmt.Errorf("помилка при створенні styles: %w", err)
	}
	styles.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" office:version="1.2">
 <office:styles/>
</office:document-styles>`))

	// Записуємо settings.xml
	settings, err := zw.Create("settings.xml")
	if err != nil {
		return fmt.Errorf("помилка при створенні settings: %w", err)
	}
	settings.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-settings xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" office:version="1.2">
 <office:settings/>
</office:document-settings>`))

	// Записуємо content.xml зі даними
	content, err := zw.Create("content.xml")
	if err != nil {
		return fmt.Errorf("помилка при створенні content: %w", err)
	}

	contentXML := generateODSContent(results)
	content.Write([]byte(contentXML))

	// Закриваємо ZIP архів та записуємо у файл
	zw.Close()

	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("помилка при записі файлу: %w", err)
	}

	log.Printf("Результати успішно записані у %s (ODS)\n", outputFile)
	return nil
}

// generateODSContent генерує XML вміст для ODS файлу
func generateODSContent(results []model.ResultRecord) string {
	var rows strings.Builder
	rows.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" 
    xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
    xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
    xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
    xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
    xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0"
    office:version="1.2">
 <office:automatic-styles/>
 <office:body>
  <office:spreadsheet>
   <table:table table:name="Sheet1" table:style-name="ta1">`)

	// Записуємо заголовок
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
	rows.WriteString("\n    <table:table-row>")
	for _, header := range headers {
		rows.WriteString("\n     <table:table-cell office:value-type=\"string\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(escapeXML(header))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")
	}
	rows.WriteString("\n    </table:table-row>")

	// Записуємо дані
	for idx, result := range results {
		rows.WriteString("\n    <table:table-row>")

		// 1. №п (число)
		rows.WriteString("\n     <table:table-cell office:value-type=\"float\" office:value=\"")
		rows.WriteString(fmt.Sprintf("%d", idx+1))
		rows.WriteString("\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(fmt.Sprintf("%d", idx+1))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")

		// 2. Постачальник (Name)
		rows.WriteString("\n     <table:table-cell office:value-type=\"string\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(escapeXML(result.Name))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")

		// 3. Дата
		rows.WriteString("\n     <table:table-cell office:value-type=\"string\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(escapeXML(result.Date))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")

		// 4. Дт рах.311 Сума
		rows.WriteString("\n     <table:table-cell office:value-type=\"float\" office:value=\"")
		rows.WriteString(fmt.Sprintf("%.2f", result.Sum))
		rows.WriteString("\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(fmt.Sprintf("%.2f", result.Sum))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")

		// 5. Оборот по Дт (Account)
		rows.WriteString("\n     <table:table-cell office:value-type=\"string\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(escapeXML(result.Account))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")

		// 6-11. Рахунки (313, 63, 641, 641.1, 651, 94) - порожні комірки
		for i := 0; i < 6; i++ {
			rows.WriteString("\n     <table:table-cell office:value-type=\"string\">")
			rows.WriteString("\n      <text:p></text:p>")
			rows.WriteString("\n     </table:table-cell>")
		}

		// 12. Оборот по Кт (Counterparty)
		rows.WriteString("\n     <table:table-cell office:value-type=\"string\">")
		rows.WriteString("\n      <text:p>")
		rows.WriteString(escapeXML(result.Counterparty))
		rows.WriteString("</text:p>")
		rows.WriteString("\n     </table:table-cell>")

		rows.WriteString("\n    </table:table-row>")
	}

	rows.WriteString(`
   </table:table>
  </office:spreadsheet>
 </office:body>
</office:document-content>`)

	return rows.String()
}

// escapeXML екранує спеціальні символи для XML
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
