package parsers

import (
	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/parsers/csv"
	"bank-analyzer/internal/parsers/pdf"
	xlsxparser "bank-analyzer/internal/parsers/xlsx"
	"fmt"
)

type Registry struct {
	parsers []Parser
}

func NewRegistry(cfg mappings.MappingsConfig) *Registry {
	r := &Registry{}
	r.Register(csv.NewPrivatBankParser(cfg))
	r.Register(xlsxparser.NewPrivatBankXLSXParser(cfg))
	r.Register(pdf.NewParser())
	return r
}

func (r *Registry) Register(p Parser) {
	r.parsers = append(r.parsers, p)
}

func (r *Registry) Detect(filepath string) (Parser, error) {
	for _, parser := range r.parsers {
		ok, err := parser.CanParse(filepath)
		if err != nil {
			continue
		}
		if ok {
			return parser, nil
		}
	}
	return nil, fmt.Errorf("невідомий формат файлу: %s", filepath)
}

func (r *Registry) ParseFile(filepath string) ([]*models.Transaction, error) {
	parser, err := r.Detect(filepath)
	if err != nil {
		return nil, err
	}
	return parser.Parse(filepath)
}
