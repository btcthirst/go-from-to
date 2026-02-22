package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Report struct {
	Transactions []*Transaction
	Period       DateRange
	TotalIncome  decimal.Decimal
	TotalExpense decimal.Decimal
	NetBalance   decimal.Decimal
	ByCategory   map[string]CategorySummary
	ByMonth      map[string]MonthSummary
}

type CategorySummary struct {
	Category string
	Count    int
	Total    decimal.Decimal
	Percent  float64
}

type MonthSummary struct {
	Month   string
	Income  decimal.Decimal
	Expense decimal.Decimal
}

type DateRange struct {
	From time.Time
	To   time.Time
}
