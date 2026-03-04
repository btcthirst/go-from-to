package categorizer

import (
	"fmt"
	"strings"

	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"
)

// ProviderResolver заповнює поле Provider у транзакціях після категоризації.
type ProviderResolver struct {
	// providers — довідник кодів для категорії "Внески": код → ПІБ.
	providers  map[string]string
	categories mappings.CategoriesConfig
}

// NewProviderResolver створює резолвер з конфігурації.
func NewProviderResolver(cfg mappings.ProvidersConfig, cats mappings.CategoriesConfig) *ProviderResolver {
	return &ProviderResolver{
		providers:  cfg.Providers,
		categories: cats,
	}
}

// Resolve заповнює tx.Provider на основі tx.Category.
// Якщо ProviderManual == true — не перезаписує (ручна зміна має пріоритет).
func (r *ProviderResolver) Resolve(tx *models.Transaction) {
	if tx.ProviderManual {
		return
	}

	switch tx.Category {
	case "Комунальні":
		tx.Provider = tx.Counterparty
	case "Внески":
		tx.Provider = r.resolveVnesky(tx)
	default:
		// для всіх інших — беремо provider з конфігурації якщо є
		if rule, ok := r.categories[tx.Category]; ok && rule.Provider != "" {
			tx.Provider = rule.Provider
		}
	}
}

// ResolveAll заповнює Provider для всіх транзакцій у списку.
func (r *ProviderResolver) ResolveAll(txs []*models.Transaction) {
	for _, tx := range txs {
		r.Resolve(tx)
	}
}

// resolveVnesky шукає код з довідника в Description+Counterparty.
// Формат результату: "ПІБ(код)", наприклад "Пупкін А.Ф(40160)".
// Повертає порожній рядок якщо жоден код не знайдено.
func (r *ProviderResolver) resolveVnesky(tx *models.Transaction) string {
	haystack := normalize(tx.Description + " " + tx.Counterparty)
	for code, name := range r.providers {
		if strings.Contains(haystack, normalize(code)) {
			return fmt.Sprintf("%s(%s)", name, code)
		}
	}
	return ""
}
