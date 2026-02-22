// Package parsers provides functions to parse Go source code and extract information about types, functions, and other constructs.
package parsers

import "bank-analyzer/internal/models"

type Parser interface {
	Name() string

	CanParse(filepath string) (bool, error)

	Parse(filepath string) ([]*models.Transaction, error)
}
