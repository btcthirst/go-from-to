// Package csv provides parsers for CSV files from various sources, such as banks, financial institutions, etc.
package csv

import (
	"bank-analyzer/internal/models"
	"encoding/csv"
	"os"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
)

type PrivatBankParser struct{}

func NewPrivatBankParser() *PrivatBankParser {
	return &PrivatBankParser{}
}

func (p *PrivatBankParser) Name() string {
	return "PrivatBank CSV Parser"
}

func (p *PrivatBankParser) CanParse(filepath string) (bool, error) {
	if !strings.HasSuffix(strings.ToLower(filepath), ".csv") {
		return false, nil
	}
	// Read the ?first row to check if it contains expected headers
	f, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	f.Close()
	// PB uses Windows-1251 encoding, but we can assume that if it's a CSV file, it's likely to be from PB
	reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
	reader.Comma = ';' // PB uses semicolon as a separator
	header, err := reader.Read()
	if err != nil {
		return false, err
	}
	return containsAll(header, []string{"Дата", "Призначення", "Сумма", "Валюта", "Категорія"}), nil
}

func (p *PrivatBankParser) Parse(filepath string) ([]*models.Transaction, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
	reader.Comma = ';' // PB uses semicolon as a separator
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	header := records[0]
	idx := buildIndex(header)

	var tsx []*models.Transaction
	for _, row := range records[1:] {
		if len(row) < 5 {
			continue // skip invalid rows
		}
		date, err := time.Parse("02.01.2006", row[idx["Дата"]])
		if err != nil {
			continue // skip rows with invalid date
		}
		amtStr := strings.ReplaceAll(row[idx["Сумма"]], " ", "")
		amt, err := decimal.NewFromString(strings.ReplaceAll(amtStr, ",", "."))
		if err != nil {
			continue // skip rows with invalid amount
		}
		txType := models.Credit
		if amt.IsNegative() {
			txType = models.Debit
			amt = amt.Abs()
		}
		tsx = append(tsx, &models.Transaction{
			Date:         date,
			Amount:       amt,
			Type:         txType,
			Currency:     row[idx["Валюта"]],
			Description:  row[idx["Призначення"]],
			Counterparty: row[idx["Контрагент"]],
			BankSource:   "",
		})
	}
	return tsx, nil
}

// Helper functions
// need to check functions below for correctness and efficiency, especially containsAll which is O(n*m)
func containsAll(slice []string, items []string) bool {
	for _, item := range items {
		found := false
		for _, s := range slice {
			if s == item {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func buildIndex(header []string) map[string]int {
	idx := make(map[string]int)
	for i, h := range header {
		idx[h] = i
	}
	return idx
}
