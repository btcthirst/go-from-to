package reader

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// SimpleODSReader reads ODS files
type SimpleODSReader struct {
	Sheets map[string][][]string
}

// OpenODS opens an ODS file and parses its content
func OpenODS(filename string) (*SimpleODSReader, error) {
	reader := &SimpleODSReader{
		Sheets: make(map[string][][]string),
	}

	// Open ZIP archive
	zipFile, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening ODS file: %w", err)
	}
	defer zipFile.Close()

	// Find content.xml
	var contentData []byte
	for _, file := range zipFile.File {
		if file.Name == "content.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening content.xml: %w", err)
			}

			contentData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("error reading content.xml: %w", err)
			}
			break
		}
	}

	if len(contentData) == 0 {
		return nil, fmt.Errorf("content.xml not found in archive")
	}

	// Parse XML content
	if err := reader.parseODSSimple(contentData); err != nil {
		return nil, fmt.Errorf("error parsing ODS: %w", err)
	}

	return reader, nil
}

// parseODSSimple is a fallback parser using simple XML handling
func (r *SimpleODSReader) parseODSSimple(data []byte) error {
	content := string(data)

	decoder := xml.NewDecoder(strings.NewReader(content))

	var currentSheet string
	var currentRow []string
	var inCell bool
	var cellDepth int
	var cellText strings.Builder
	var rowCount int

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding XML: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			// Check for table element
			if se.Name.Local == "table" {
				for _, attr := range se.Attr {
					if attr.Name.Local == "name" {
						currentSheet = attr.Value
						r.Sheets[currentSheet] = [][]string{}
						rowCount = 0
						break
					}
				}
			}

			// Check for table-row element
			if se.Name.Local == "table-row" && currentSheet != "" {
				currentRow = []string{}
			}

			// Check for table-cell element
			if se.Name.Local == "table-cell" {
				inCell = true
				cellDepth = 0
				cellText.Reset()
			}

			if inCell {
				cellDepth++
			}

		case xml.EndElement:
			// End of table
			if se.Name.Local == "table" && currentSheet != "" {
				currentSheet = ""
			}

			// End of row
			if se.Name.Local == "table-row" && currentSheet != "" {
				if len(currentRow) > 0 {
					if sheetData, exists := r.Sheets[currentSheet]; exists {
						r.Sheets[currentSheet] = append(sheetData, currentRow)
						rowCount++
					}
				}
				currentRow = []string{}
			}

			// End of cell
			if se.Name.Local == "table-cell" && inCell {
				currentRow = append(currentRow, strings.TrimSpace(cellText.String()))
				inCell = false
				cellText.Reset()
				cellDepth = 0
			}

			if inCell && cellDepth > 0 {
				cellDepth--
			}

		case xml.CharData:
			if inCell {
				cellText.Write(se)
			}
		}
	}

	return nil
}

// GetSheet returns sheet data
func (r *SimpleODSReader) GetSheet(sheetName string) ([][]string, error) {
	if sheet, exists := r.Sheets[sheetName]; exists {
		return sheet, nil
	}

	// If sheet not found, return first sheet
	for _, sheet := range r.Sheets {
		return sheet, nil
	}

	return nil, fmt.Errorf("no sheets found in document")
}

// GetAllSheetNames returns all sheet names
func (r *SimpleODSReader) GetAllSheetNames() []string {
	var names []string
	for name := range r.Sheets {
		names = append(names, name)
	}
	return names
}
