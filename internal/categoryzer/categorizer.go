// Package categoryzer provides a categorizer for categorizing items based on certain criteria.
package categoryzer

// NewEmptyCategorizer повертає категоризатор без правил (fallback).
// Використовується як fallback коли categories.yaml недоступний —
// застосунок запускається, всі транзакції отримують категорію "Uncategorized".
func NewEmptyCategorizer() *RulesCategorizer {
	return &RulesCategorizer{
		categories:     make(CategoryConfig),
		debitKeywords:  make(map[string]keywordEntry),
		creditKeywords: make(map[string]keywordEntry),
	}
}
