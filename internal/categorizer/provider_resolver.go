package categorizer

import (
	"fmt"
	"strings"

	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"
)

// categoryProvider — фіксований провайдер для простих категорій (1-до-1).
var categoryProvider = map[string]string{
	"Військовий":    "військовий збір 5%",
	"ПДФО":          "пдфо 18%",
	"ЄСВ":           "єсв 22%",
	"Комісія банку": "комісія банку",
	"Зарплата":      "кошти на картку(з/п)",
}

// communalKeywords — підпостачальники для категорії "Комунальні".
// Ключ — підрядок для пошуку (lowercase), значення — відображувана назва.
var communalKeywords = map[string]string{
	"вувкг": "вувкг",
	"парки": "парки",
	"хоек":  "хоек",
}

// ProviderResolver заповнює поле Provider у транзакціях після категоризації.
type ProviderResolver struct {
	// providers — довідник кодів для категорії "Внески": код → ПІБ.
	providers map[string]string
}

// NewProviderResolver створює резолвер з конфігурації.
func NewProviderResolver(cfg mappings.ProvidersConfig) *ProviderResolver {
	return &ProviderResolver{providers: cfg.Providers}
}

// Resolve заповнює tx.Provider на основі tx.Category.
// Якщо ProviderManual == true — не перезаписує (ручна зміна має пріоритет).
func (r *ProviderResolver) Resolve(tx *models.Transaction) {
	if tx.ProviderManual {
		return
	}

	switch tx.Category {
	case "Військовий", "ПДФО", "ЄСВ", "Комісія банку", "Зарплата":
		tx.Provider = categoryProvider[tx.Category]

	case "Комунальні":
		tx.Provider = r.resolveKommunal(tx)

	case "Внески":
		tx.Provider = r.resolveVnesky(tx)
	}
}

// ResolveAll заповнює Provider для всіх транзакцій у списку.
func (r *ProviderResolver) ResolveAll(txs []*models.Transaction) {
	for _, tx := range txs {
		r.Resolve(tx)
	}
}

// resolveKommunal шукає підпостачальника в Description+Counterparty.
// Якщо жоден ключ не знайдено — повертає назву категорії як fallback.
func (r *ProviderResolver) resolveKommunal(tx *models.Transaction) string {
	haystack := strings.ToLower(tx.Description + " " + tx.Counterparty)
	for key, name := range communalKeywords {
		if strings.Contains(haystack, key) {
			return name
		}
	}
	return tx.Category
}

// resolveVnesky шукає код з довідника в Description+Counterparty.
// Формат результату: "ПІБ(код)", наприклад "Пупкін А.Ф(40160)".
// Повертає порожній рядок якщо жоден код не знайдено.
func (r *ProviderResolver) resolveVnesky(tx *models.Transaction) string {
	haystack := strings.ToLower(tx.Description + " " + tx.Counterparty)
	for code, name := range r.providers {
		if strings.Contains(haystack, strings.ToLower(code)) {
			return fmt.Sprintf("%s(%s)", name, code)
		}
	}
	return ""
}
