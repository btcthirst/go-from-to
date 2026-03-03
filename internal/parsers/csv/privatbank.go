// Package csv implements parsers for CSV bank statements.
package csv

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/parsers/utils"

	"golang.org/x/text/encoding/charmap"
)

const privatbankCSVKey = "privatbank_csv"

// PrivatBankParser parses PrivatBank CSV statements.
// The CSV file uses Windows-1251 encoding and semicolon as delimiter.
type PrivatBankParser struct {
	mapping mappings.ParserMapping
}

// NewPrivatBankParser creates a parser using the provided column mappings.
// Panics if the mapping key is not found — see MappingsConfig.MustGetParser.
func NewPrivatBankParser(cfg mappings.MappingsConfig) *PrivatBankParser {
	return &PrivatBankParser{mapping: cfg.MustGetParser(privatbankCSVKey)}
}

func (p *PrivatBankParser) Name() string { return "ПриватБанк CSV" }

// requiredCSVHeaders is the minimum set of headers needed to identify the format.
var requiredCSVHeaders = []string{"ЄДРПОУ", "МФО", "Рахунок", "Дата операції", "Сума"}

func (p *PrivatBankParser) CanParse(filepath string) (bool, error) {
	if !strings.HasSuffix(strings.ToLower(filepath), ".csv") {
		return false, nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return false, nil
	}

	log.Printf("[PrivatBankCSV] CanParse: headers = %v", header)

	normalized := normalizeHeaders(header)
	for _, required := range requiredCSVHeaders {
		if !normalized[strings.ToLower(strings.TrimSpace(required))] {
			log.Printf("[PrivatBankCSV] CanParse: header %q not found", required)
			return false, nil
		}
	}
	return true, nil
}

func (p *PrivatBankParser) Parse(filepath string) ([]*models.Transaction, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV read error: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("file is empty or contains only headers")
	}

	header := records[0]
	log.Printf("[PrivatBankCSV] Parse: headers = %v", header)

	idx := buildIndex(header)

	dateCol := utils.FirstMatch(idx, p.mapping.Get("date"))
	amountCol := utils.FirstMatch(idx, p.mapping.Get("amount"))
	currencyCol := utils.FirstMatch(idx, p.mapping.Get("currency"))
	descCol := utils.FirstMatch(idx, p.mapping.Get("description"))
	counterpartyCol := utils.FirstMatch(idx, p.mapping.Get("counterparty"))
	ibanCol := utils.FirstMatch(idx, p.mapping.Get("iban"))
	edropuCol := utils.FirstMatch(idx, p.mapping.Get("edrpou"))
	docNumCol := utils.FirstMatch(idx, p.mapping.Get("doc_num"))

	log.Printf("[PrivatBankCSV] Parse: columns: date=%d amount=%d currency=%d desc=%d counterparty=%d iban=%d",
		dateCol, amountCol, currencyCol, descCol, counterpartyCol, ibanCol)

	if dateCol < 0 {
		return nil, fmt.Errorf("date column not found")
	}
	if amountCol < 0 {
		return nil, fmt.Errorf("amount column not found")
	}

	var txs []*models.Transaction
	skipped := 0

	for rowNum, row := range records[1:] {
		if len(row) == 0 || isEmptyRow(row) {
			continue
		}
		tx, err := parseCSVRow(row, rowNum+2, dateCol, amountCol, currencyCol, descCol, counterpartyCol, ibanCol, edropuCol, docNumCol)
		if err != nil {
			log.Printf("[PrivatBankCSV] skipping row %d: %v", rowNum+2, err)
			skipped++
			continue
		}
		tx.BankSource = "privatbank"
		txs = append(txs, tx)
	}

	log.Printf("[PrivatBankCSV] Parse: imported=%d, skipped=%d", len(txs), skipped)

	if len(txs) == 0 {
		return nil, fmt.Errorf("no transactions found (skipped: %d)", skipped)
	}
	return txs, nil
}

func parseCSVRow(row []string, rowNum, dateCol, amountCol, currencyCol, descCol, counterpartyCol, ibanCol, edropuCol, docNumCol int) (*models.Transaction, error) {
	dateRaw := utils.SafeGet(row, dateCol)
	date, err := utils.ParseDate(dateRaw)
	if err != nil {
		return nil, fmt.Errorf("row %d: date %q: %w", rowNum, dateRaw, err)
	}

	amountRaw := utils.SafeGet(row, amountCol)
	amt, err := utils.ParseDecimal(amountRaw)
	if err != nil {
		return nil, fmt.Errorf("row %d: amount %q: %w", rowNum, amountRaw, err)
	}
	if amt.IsZero() {
		return nil, fmt.Errorf("row %d: amount is zero", rowNum)
	}

	txType := models.Credit
	if amt.IsNegative() {
		txType = models.Debit
		amt = amt.Abs()
	}

	currency := utils.SafeGet(row, currencyCol)
	if currency == "" {
		currency = "UAH"
	}

	raw := make(map[string]string)
	if v := utils.SafeGet(row, docNumCol); v != "" {
		raw["doc_num"] = v
	}
	if v := utils.SafeGet(row, edropuCol); v != "" {
		raw["edrpou"] = v
	}

	return &models.Transaction{
		Date:         date,
		Amount:       amt,
		Type:         txType,
		Currency:     currency,
		Description:  utils.SafeGet(row, descCol),
		Counterparty: utils.SafeGet(row, counterpartyCol),
		IBAN:         utils.SafeGet(row, ibanCol),
		Raw:          raw,
	}, nil
}

// --- Shared CSV helpers ---

func normalizeHeaders(header []string) map[string]bool {
	m := make(map[string]bool, len(header))
	for _, h := range header {
		m[strings.ToLower(strings.TrimSpace(h))] = true
	}
	return m
}

func buildIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
