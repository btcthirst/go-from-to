// Package xlsx provides functions to parse XLSX files and extract data from them.
package xlsx

import "bank-analyzer/internal/models"

type XLSXParser struct {
}

func NewParser() *XLSXParser {
	return &XLSXParser{}
}

func (p *XLSXParser) Name() string {
	return "XLSX Parser"
}

func (p *XLSXParser) CanParse(filepath string) (bool, error) {
	// Implement logic to check if the file is an XLSX and can be parsed
	return false, nil
}

func (p *XLSXParser) Parse(filepath string) ([]*models.Transaction, error) {
	// Implement logic to parse the XLSX file and extract transactions
	return nil, nil
}
