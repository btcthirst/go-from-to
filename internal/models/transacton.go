// Package models contains the data structures and types used in the application, such as transactions, accounts, and other related entities. It serves as a central place for defining the core data models that are used throughout the application.
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
	Balance      decimal.Decimal
	BankSource   string
	Raw          map[string]string
}

func (t *Transaction) NetAmount() decimal.Decimal {
	if t.Type == Credit {
		return t.Amount
	}
	return t.Amount.Neg()
}
