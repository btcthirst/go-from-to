// Package models contains the data structures and types used in the application.
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionType int

const (
	Debit TransactionType = iota
	Credit
)

type Transaction struct {
	ID           string
	Date         time.Time
	Amount       decimal.Decimal
	Type         TransactionType
	Currency     string
	Description  string
	Counterparty string
	IBAN         string
	Category     string
	// Provider — ідентифікований постачальник/платник для звіту.
	// Заповнюється після категоризації через ProviderResolver.
	// Приклади: "ПДФО 18%", "вувкг", "Пупкін А.Ф(40160)"
	Provider string
	// ProviderManual — true якщо Provider встановлено вручну користувачем.
	// ProviderResolver не перезаписує поле якщо цей прапор встановлено.
	ProviderManual bool
	Balance        decimal.Decimal
	BankSource     string
	Raw            map[string]string
}

func (t *Transaction) NetAmount() decimal.Decimal {
	if t.Type == Credit {
		return t.Amount
	}
	return t.Amount.Neg()
}

// ToDTO конвертує Transaction у TransactionDTO для збереження/передачі.
func (t *Transaction) ToDTO() TransactionDTO {
	amt, _ := t.Amount.Float64()
	return TransactionDTO{
		ID:           t.ID,
		Date:         t.Date.Format("02.01.2006"),
		Amount:       amt,
		Type:         t.Type,
		Currency:     t.Currency,
		Counterparty: t.Counterparty,
		Category:     t.Category,
		Provider:     t.Provider,
	}
}

// TransactionDTO — спрощена структура для збереження/експорту транзакцій.
type TransactionDTO struct {
	ID           string
	Date         string
	Amount       float64
	Type         TransactionType
	Currency     string
	Counterparty string
	Category     string
	Provider     string
}

// TypeLabel повертає людиночитабельну назву типу транзакції.
func (d TransactionDTO) TypeLabel() string {
	if d.Type == Debit {
		return "Витрата"
	}
	return "Надходження"
}

// ToTransactions конвертує зріз Transaction у зріз TransactionDTO.
func ToTransactions(txs []*Transaction) []TransactionDTO {
	result := make([]TransactionDTO, len(txs))
	for i, tx := range txs {
		result[i] = tx.ToDTO()
	}
	return result
}
