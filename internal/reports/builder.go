// Package reports provides functionality to build and generate reports based on collected data.
package reports

import (
	"bank-analyzer/internal/models"

	"github.com/shopspring/decimal"
)

// BuildReport агрегує транзакції у models.Report.
// Приймає вже відфільтрований зріз транзакцій — фільтрація по даті
// залишається відповідальністю UI або викликача.
func BuildReport(txs []*models.Transaction, opts BuildOptions) *models.Report {
	report := &models.Report{
		Transactions: txs,
		ByCategory:   make(map[string]models.CategorySummary),
		ByMonth:      make(map[string]models.MonthSummary),
	}

	if len(txs) == 0 {
		return report
	}

	report.Period.From = txs[0].Date
	report.Period.To = txs[0].Date

	var totalExpense decimal.Decimal

	for _, tx := range txs {
		if tx.Date.Before(report.Period.From) {
			report.Period.From = tx.Date
		}
		if tx.Date.After(report.Period.To) {
			report.Period.To = tx.Date
		}

		if tx.Type == models.Credit {
			report.TotalIncome = report.TotalIncome.Add(tx.Amount)
		} else {
			report.TotalExpense = report.TotalExpense.Add(tx.Amount)
			totalExpense = totalExpense.Add(tx.Amount)
		}

		if opts.IncludeCategory {
			cs := report.ByCategory[tx.Category]
			cs.Category = tx.Category
			cs.Count++
			cs.Total = cs.Total.Add(tx.Amount)
			report.ByCategory[tx.Category] = cs
		}

		if opts.IncludeMonthly {
			key := tx.Date.Format("2006-01")
			ms := report.ByMonth[key]
			ms.Month = key
			if tx.Type == models.Credit {
				ms.Income = ms.Income.Add(tx.Amount)
			} else {
				ms.Expense = ms.Expense.Add(tx.Amount)
			}
			report.ByMonth[key] = ms
		}
	}

	report.NetBalance = report.TotalIncome.Sub(report.TotalExpense)

	if opts.IncludeCategory && !totalExpense.IsZero() {
		for key, cs := range report.ByCategory {
			pct, _ := cs.Total.Div(totalExpense).Mul(decimal.NewFromInt(100)).Float64()
			cs.Percent = pct
			report.ByCategory[key] = cs
		}
	}

	return report
}

// BuildOptions керує тим які секції звіту заповнюються.
// Відповідає чекбоксам у UI екрані звіту.
type BuildOptions struct {
	IncludeCategory bool
	IncludeMonthly  bool
}
