// Package pdf provides a parser for PDF files.
package pdf

import "bank-analyzer/internal/models"

type PDFParser struct {
}

func NewParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Name() string {
	return "PDF Parser"
}

func (p *PDFParser) CanParse(filepath string) (bool, error) {
	// Implement logic to check if the file is a PDF and can be parsed
	return false, nil
}

func (p *PDFParser) Parse(filepath string) ([]*models.Transaction, error) {
	// Implement logic to parse the PDF file and extract transactions
	return nil, nil
}
