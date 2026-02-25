package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"strings"
	"time"
)

const mimeType = "application/vnd.oasis.opendocument.spreadsheet"

type Row []string

func WriteODS(path string, sheetName string, rows []Row) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// --- 1. mimetype (STORE, перший)
	if err := writeMimeType(zw); err != nil {
		return err
	}

	if err := writeDir(zw, "META-INF/"); err != nil {
		return err
	}

	// --- 2. content.xml
	if err := writeDeflated(zw, "content.xml", buildContent(sheetName, rows)); err != nil {
		return err
	}

	// --- 3. styles.xml
	if err := writeDeflated(zw, "styles.xml", stylesXML); err != nil {
		return err
	}

	// --- 4. meta.xml
	if err := writeDeflated(zw, "meta.xml", buildMeta()); err != nil {
		return err
	}

	// --- 5. manifest.xml
	if err := writeDeflated(zw, "META-INF/manifest.xml", manifestXML); err != nil {
		return err
	}

	return nil
}

func writeMimeType(zw *zip.Writer) error {
	data := []byte(mimeType)

	h := &zip.FileHeader{
		Name:               "mimetype",
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE(data),
		CompressedSize64:   uint64(len(data)),
		UncompressedSize64: uint64(len(data)),
	}

	w, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func writeDir(zw *zip.Writer, name string) error {
	h := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	}
	_, err := zw.CreateHeader(h)
	return err
}

func writeDeflated(zw *zip.Writer, name string, data string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(data))
	return err
}

func buildContent(sheet string, rows []Row) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<office:document-content
xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
office:version="1.4">`)

	sb.WriteString(`<office:automatic-styles/>`)
	sb.WriteString(`<office:body><office:spreadsheet>`)
	sb.WriteString(`<table:table table:name="`)
	xml.EscapeText(&sb, []byte(sheet))
	sb.WriteString(`">`)

	for _, row := range rows {
		sb.WriteString(`<table:table-row>`)
		for _, cell := range row {
			sb.WriteString(`<table:table-cell office:value-type="string"><text:p>`)
			xml.EscapeText(&sb, []byte(cell))
			sb.WriteString(`</text:p></table:table-cell>`)
		}
		sb.WriteString(`</table:table-row>`)
	}

	sb.WriteString(`</table:table>`)
	sb.WriteString(`</office:spreadsheet></office:body>`)
	sb.WriteString(`</office:document-content>`)

	return sb.String()
}

func buildMeta() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta
xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
xmlns:dc="http://purl.org/dc/elements/1.1/"
office:version="1.4">
<office:meta>
<meta:generator>minimal-ods-writer</meta:generator>
<dc:date>%s</dc:date>
</office:meta>
</office:document-meta>`, time.Now().Format(time.RFC3339))
}

const stylesXML = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-styles
xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
office:version="1.4">
<office:styles/>
</office:document-styles>`

const manifestXML = `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest
xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0"
manifest:version="1.4">

<manifest:file-entry
 manifest:full-path="/"
 manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/>

<manifest:file-entry
 manifest:full-path="META-INF/"
 manifest:media-type=""/>

<manifest:file-entry
 manifest:full-path="mimetype"
 manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/>

<manifest:file-entry
 manifest:full-path="content.xml"
 manifest:media-type="text/xml"/>

<manifest:file-entry
 manifest:full-path="styles.xml"
 manifest:media-type="text/xml"/>

<manifest:file-entry
 manifest:full-path="meta.xml"
 manifest:media-type="text/xml"/>

</manifest:manifest>`

func main() {
	rows := []Row{
		{"ID", "Name", "Amount"},
		{"1", "Test", "100"},
	}

	err := WriteODS("test.ods", "Sheet1", rows)
	if err != nil {
		log.Fatal(err)
	}
}
