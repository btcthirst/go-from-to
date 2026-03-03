// Package xlsx implements parsers for XLSX bank statements.
package xlsx

import (
	"fmt"
	"log"
	"strings"

	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/parsers/utils"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

const privatbankXLSXKey = "privatbank_xlsx"

// PrivatBankXLSXParser parses PrivatBank XLSX statements.
type PrivatBankXLSXParser struct {
	mapping mappings.ParserMapping
}

// NewPrivatBankXLSXParser creates a parser using the provided column mappings.
// Panics if the mapping key is not found — see MappingsConfig.MustGetParser.
func NewPrivatBankXLSXParser(cfg mappings.MappingsConfig) *PrivatBankXLSXParser {
	return &PrivatBankXLSXParser{mapping: cfg.MustGetParser(privatbankXLSXKey)}
}

func (p *PrivatBankXLSXParser) Name() string { return "ПриватБанк XLSX" }

// requiredXLSXHeaders contains sets of headers used to identify the format.
// Any one matching set is sufficient.
var requiredXLSXHeaders = [][]string{
	{"Дата проводки", "Сума", "Призначення платежу"},
	{"Дата проводки", "Сума в валюті рахунку", "Призначення платежу"},
	{"Дата", "Сума", "Призначення"},
	{"Дата", "Сума", "Призначення платежу"},
}

func (p *PrivatBankXLSXParser) CanParse(filepath string) (bool, error) {
	if !strings.HasSuffix(strings.ToLower(filepath), ".xlsx") {
		return false, nil
	}
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		return false, nil
	}
	defer f.Close()

	sheet, err := firstSheet(f)
	if err != nil {
		return false, nil
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return false, nil
	}

	_, found := findHeaderRow(rows)
	return found, nil
}

func (p *PrivatBankXLSXParser) Parse(filepath string) ([]*models.Transaction, error) {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	sheet, err := firstSheet(f)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("could not read rows: %w", err)
	}

	headerIdx, found := findHeaderRow(rows)
	if !found {
		return nil, fmt.Errorf("header row not found")
	}

	header := rows[headerIdx]
	idx := buildXLSXIndex(header)

	dateCol := utils.FirstMatch(idx, p.mapping.Get("date"))
	amountCol := utils.FirstMatch(idx, p.mapping.Get("amount"))
	currencyCol := utils.FirstMatch(idx, p.mapping.Get("currency"))
	descCol := utils.FirstMatch(idx, p.mapping.Get("description"))
	counterpartyCol := utils.FirstMatch(idx, p.mapping.Get("counterparty"))
	ibanCol := utils.FirstMatch(idx, p.mapping.Get("iban"))
	balanceCol := utils.FirstMatch(idx, p.mapping.Get("balance"))
	debitCol := utils.FirstMatch(idx, p.mapping.Get("debit"))
	creditCol := utils.FirstMatch(idx, p.mapping.Get("credit"))

	if dateCol < 0 {
		return nil, fmt.Errorf("date column not found")
	}

	var txs []*models.Transaction
	skipped := 0

	for rowNum, row := range rows[headerIdx+1:] {
		if len(row) == 0 || isXLSXEmptyRow(row) {
			continue
		}
		if isSummaryRow(row) {
			log.Printf("[PrivatBankXLSX] stopping at summary row %d", rowNum+headerIdx+2)
			break
		}

		tx, err := parseXLSXRow(row, rowNum+headerIdx+2, dateCol, amountCol, debitCol, creditCol, currencyCol, descCol, counterpartyCol, ibanCol, balanceCol)
		if err != nil {
			log.Printf("[PrivatBankXLSX] skipping row %d: %v | %v", rowNum+headerIdx+2, err, row)
			skipped++
			continue
		}
		tx.BankSource = "privatbank"
		txs = append(txs, tx)
	}

	if len(txs) == 0 {
		return nil, fmt.Errorf("no transactions found (skipped: %d)", skipped)
	}
	return txs, nil
}

// --- Helpers ---

func firstSheet(f *excelize.File) (string, error) {
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", fmt.Errorf("file contains no sheets")
	}
	return sheets[0], nil
}

func findHeaderRow(rows [][]string) (int, bool) {
	limit := 20
	if len(rows) < limit {
		limit = len(rows)
	}
	for i, row := range rows[:limit] {
		if matchesAnyHeaderSet(row) {
			return i, true
		}
	}
	return -1, false
}

func matchesAnyHeaderSet(row []string) bool {
	normalized := make(map[string]bool, len(row))
	for _, cell := range row {
		normalized[strings.ToLower(strings.TrimSpace(cell))] = true
	}
	for _, headerSet := range requiredXLSXHeaders {
		match := true
		for _, h := range headerSet {
			if !normalized[strings.ToLower(h)] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func buildXLSXIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}

func isXLSXEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func isSummaryRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(row[0]))
	for _, marker := range []string{"разом", "всього", "підсумок", "итого", "total"} {
		if strings.HasPrefix(first, marker) {
			return true
		}
	}
	return false
}

func parseXLSXRow(row []string, rowNum, dateCol, amountCol, debitCol, creditCol, currencyCol, descCol, counterpartyCol, ibanCol, balanceCol int) (*models.Transaction, error) {
	dateRaw := utils.SafeGet(row, dateCol)
	date, err := utils.ParseDate(dateRaw)
	if err != nil {
		return nil, fmt.Errorf("date %q: %w", dateRaw, err)
	}

	var amount decimal.Decimal
	var txType models.TransactionType

	if debitCol >= 0 || creditCol >= 0 {
		debit, _ := utils.ParseDecimal(utils.SafeGet(row, debitCol))
		credit, _ := utils.ParseDecimal(utils.SafeGet(row, creditCol))
		if !credit.IsZero() {
			amount, txType = credit, models.Credit
		} else if !debit.IsZero() {
			amount, txType = debit, models.Debit
		} else {
			return nil, fmt.Errorf("row %d: both debit and credit are zero", rowNum)
		}
	} else if amountCol >= 0 {
		amt, err := utils.ParseDecimal(utils.SafeGet(row, amountCol))
		if err != nil {
			return nil, fmt.Errorf("amount: %w", err)
		}
		if amt.IsNegative() {
			amount, txType = amt.Abs(), models.Debit
		} else {
			amount, txType = amt, models.Credit
		}
	} else {
		return nil, fmt.Errorf("row %d: amount column not found", rowNum)
	}

	currency := utils.SafeGet(row, currencyCol)
	if currency == "" {
		currency = "UAH"
	}

	var balance decimal.Decimal
	if balanceCol >= 0 {
		balance, _ = utils.ParseDecimal(utils.SafeGet(row, balanceCol))
	}

	raw := make(map[string]string, len(row))
	for i, cell := range row {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		raw[colName] = cell
	}

	return &models.Transaction{
		Date:         date,
		Amount:       amount,
		Type:         txType,
		Currency:     currency,
		Description:  utils.SafeGet(row, descCol),
		Counterparty: utils.SafeGet(row, counterpartyCol),
		IBAN:         utils.SafeGet(row, ibanCol),
		Balance:      balance,
		Raw:          raw,
	}, nil
}
