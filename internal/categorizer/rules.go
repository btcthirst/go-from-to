package categorizer

import (
	"log"
	"strings"
	"unicode"

	"bank-analyzer/internal/mappings"
	"bank-analyzer/internal/models"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// keywordEntry — один запис кешу: категорія + тип транзакції.
type keywordEntry struct {
	category string
	txType   string // "debit", "credit"
}

type RulesCategorizer struct {
	categories mappings.CategoriesConfig

	// debitKeywords — ключові слова для витрат (type: debit)
	debitKeywords map[string]keywordEntry

	// creditKeywords — ключові слова для надходжень (type: credit)
	creditKeywords map[string]keywordEntry
}

// NewRulesCategorizer створює категоризатор з готової конфігурації.
// Конфігурація завантажується зовні через mappings.LoadCategoriesConfig.
func NewRulesCategorizer(cfg mappings.CategoriesConfig) *RulesCategorizer {
	c := &RulesCategorizer{
		categories:     cfg,
		debitKeywords:  make(map[string]keywordEntry),
		creditKeywords: make(map[string]keywordEntry),
	}
	c.buildKeywordMaps()
	return c
}

// buildKeywordMaps будує два окремих кеші — для debit і credit транзакцій.
func (c *RulesCategorizer) buildKeywordMaps() {
	for category, rule := range c.categories {
		for _, kw := range rule.Keywords {
			key := normalize(kw)
			if key == "" {
				continue
			}
			entry := keywordEntry{category: category, txType: rule.Type}
			switch strings.ToLower(rule.Type) {
			case "debit":
				c.debitKeywords[key] = entry
			case "credit":
				c.creditKeywords[key] = entry
			default:
				log.Printf("[RulesCategorizer] категорія %q має невідомий тип %q — пропускаємо ключове слово %q",
					category, rule.Type, kw)
			}
		}
	}
}

// Categorize визначає категорію транзакції по ключових словах.
func (c *RulesCategorizer) Categorize(tx *models.Transaction) string {
	desc := normalize(tx.Description + " " + tx.Counterparty)

	var keywords map[string]keywordEntry
	switch tx.Type {
	case models.Credit:
		keywords = c.creditKeywords
	default:
		keywords = c.debitKeywords
	}

	for kw, entry := range keywords {
		if strings.Contains(desc, kw) {
			return entry.category
		}
	}

	return models.UncategorizedCategory
}

// CategorizeAll категоризує всі транзакції у списку.
// Пропускає транзакції що вже мають категорію.
func (c *RulesCategorizer) CategorizeAll(txs []*models.Transaction) {
	for _, tx := range txs {
		if tx.Category == "" || tx.Category == models.UncategorizedCategory {
			tx.Category = c.Categorize(tx)
		}
	}
}

// AddRule додає нове правило і оновлює кеші.
func (c *RulesCategorizer) AddRule(category, keyword, txType string) {
	kw := normalize(keyword)
	if kw == "" {
		return
	}

	entry := keywordEntry{category: category, txType: txType}

	switch strings.ToLower(txType) {
	case "debit":
		c.debitKeywords[kw] = entry
	case "credit":
		c.creditKeywords[kw] = entry
	default:
		log.Printf("[RulesCategorizer] AddRule: невідомий тип %q для категорії %q — правило не додано",
			txType, category)
		return
	}

	rule := c.categories[category]
	rule.Keywords = append(rule.Keywords, kw)
	if rule.Type == "" && txType != "" {
		rule.Type = txType
	}
	c.categories[category] = rule
}

// normalize приводить рядок до єдиного вигляду:
// NFKD декомпозиція + видалення діакритики + lowercase.
func normalize(s string) string {
	t := transform.Chain(
		norm.NFKD,
		runes.Remove(runes.In(unicode.Mn)),
	)
	result, _, _ := transform.String(t, strings.ToLower(s))
	return result
}
