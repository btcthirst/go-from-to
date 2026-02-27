// Package categorizer provides a categorizer for categorizing items based on certain criteria.
package categorizer

import "bank-analyzer/internal/mappings"

// NewEmptyCategorizer повертає категоризатор без правил (fallback).
// Використовується коли categories.yaml недоступний —
// застосунок запускається, всі транзакції отримують категорію "Uncategorized".
func NewEmptyCategorizer() *RulesCategorizer {
	return NewRulesCategorizer(mappings.CategoriesConfig{})
}
