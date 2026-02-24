// Package csv implements a parser for CSV bank statements, specifically tailored for ПриватБанк.
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

type PrivatBankParser struct {
	mapping mappings.ParserMapping
}

func NewPrivatBankParser(cfg mappings.MappingsConfig) *PrivatBankParser {
	m, err := cfg.ForParser(privatbankCSVKey)
	if err != nil {
		log.Printf("[PrivatBankCSV] маппінг не знайдено, використовую порожній: %v", err)
		m = mappings.ParserMapping{}
	}
	return &PrivatBankParser{mapping: m}
}

func (p *PrivatBankParser) Name() string { return "ПриватБанк CSV" }

// requiredHeaders — мінімальний набір для ідентифікації формату.
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

	log.Printf("[PrivatBankCSV] CanParse: заголовки = %v", header)

	normalized := normalizeHeaders(header)
	for _, required := range requiredCSVHeaders {
		if !normalized[strings.ToLower(strings.TrimSpace(required))] {
			log.Printf("[PrivatBankCSV] CanParse: не знайдено заголовок %q", required)
			return false, nil
		}
	}
	return true, nil
}

func (p *PrivatBankParser) Parse(filepath string) ([]*models.Transaction, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("не вдалося відкрити файл: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("помилка читання CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("файл порожній або містить лише заголовки")
	}

	header := records[0]
	log.Printf("[PrivatBankCSV] Parse: заголовки = %v", header)

	idx := buildIndex(header)

	dateCol := utils.FirstMatch(idx, p.mapping.Get("date"))
	amountCol := utils.FirstMatch(idx, p.mapping.Get("amount"))
	currencyCol := utils.FirstMatch(idx, p.mapping.Get("currency"))
	descCol := utils.FirstMatch(idx, p.mapping.Get("description"))
	counterpartyCol := utils.FirstMatch(idx, p.mapping.Get("counterparty"))
	ibanCol := utils.FirstMatch(idx, p.mapping.Get("iban"))
	edropuCol := utils.FirstMatch(idx, p.mapping.Get("edrpou"))
	docNumCol := utils.FirstMatch(idx, p.mapping.Get("doc_num"))

	log.Printf("[PrivatBankCSV] Parse: колонки: date=%d amount=%d currency=%d desc=%d counterparty=%d iban=%d",
		dateCol, amountCol, currencyCol, descCol, counterpartyCol, ibanCol)

	if dateCol < 0 {
		return nil, fmt.Errorf("не знайдено колонку дати")
	}
	if amountCol < 0 {
		return nil, fmt.Errorf("не знайдено колонку суми")
	}

	var txs []*models.Transaction
	skipped := 0

	for rowNum, row := range records[1:] {
		if len(row) == 0 || isEmptyRow(row) {
			continue
		}
		tx, err := parseCSVRow(row, rowNum+2, dateCol, amountCol, currencyCol, descCol, counterpartyCol, ibanCol, edropuCol, docNumCol)
		if err != nil {
			log.Printf("[PrivatBankCSV] пропущено рядок %d: %v", rowNum+2, err)
			skipped++
			continue
		}
		tx.BankSource = "privatbank"
		txs = append(txs, tx)
	}

	log.Printf("[PrivatBankCSV] Parse: імпортовано=%d, пропущено=%d", len(txs), skipped)

	if len(txs) == 0 {
		return nil, fmt.Errorf("не знайдено жодної транзакції (пропущено: %d)", skipped)
	}
	return txs, nil
}

func parseCSVRow(row []string, rowNum, dateCol, amountCol, currencyCol, descCol, counterpartyCol, ibanCol, edropuCol, docNumCol int) (*models.Transaction, error) {
	dateRaw := utils.SafeGet(row, dateCol)
	date, err := utils.ParseDate(dateRaw)
	if err != nil {
		return nil, fmt.Errorf("рядок %d: дата %q: %w", rowNum, dateRaw, err)
	}

	amountRaw := utils.SafeGet(row, amountCol)
	amt, err := utils.ParseDecimal(amountRaw)
	if err != nil {
		return nil, fmt.Errorf("рядок %d: сума %q: %w", rowNum, amountRaw, err)
	}
	if amt.IsZero() {
		return nil, fmt.Errorf("рядок %d: сума нульова", rowNum)
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

// --- Спільні допоміжні функції для CSV пакету ---

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
