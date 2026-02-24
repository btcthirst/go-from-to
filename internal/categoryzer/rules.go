package categoryzer

import (
	"log"
	"os"
	"strings"

	"bank-analyzer/internal/models"

	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v2"
)

type Rule struct {
	Keywords []string `yaml:"keywords"`
	MinAmt   float64  `yaml:"min_amount,omitempty"`
	MaxAmt   float64  `yaml:"max_amount,omitempty"`
	// Type визначає до яких транзакцій застосовується правило:
	// "debit"  — тільки витрати
	// "credit" — тільки надходження
	// ""       — будь-який тип (універсальне правило)
	Type string `yaml:"type,omitempty"`
}

type CategoryConfig map[string]Rule

// keywordEntry — один запис кешу: ключове слово + метадані правила.
type keywordEntry struct {
	category string
	txType   string // "debit", "credit", або "" (будь-який)
}

type RulesCategorizer struct {
	categories CategoryConfig

	// debitKeywords — ключові слова для витрат (type: debit) та універсальних (type: "")
	debitKeywords map[string]keywordEntry

	// creditKeywords — ключові слова для надходжень (type: credit) та універсальних (type: "")
	creditKeywords map[string]keywordEntry
}

func NewRulesCategorizer(configPath string) (*RulesCategorizer, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg CategoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	c := &RulesCategorizer{
		categories:     cfg,
		debitKeywords:  make(map[string]keywordEntry),
		creditKeywords: make(map[string]keywordEntry),
	}
	c.buildKeywordMaps()
	return c, nil
}

// buildKeywordMaps будує два окремих кеші — для debit і credit транзакцій.
// Універсальні правила (type: "") потрапляють в обидва кеші.
func (c *RulesCategorizer) buildKeywordMaps() {
	for category, rule := range c.categories {
		for _, kw := range rule.Keywords {
			key := normalize(kw)
			if key == "" {
				continue
			}
			entry := keywordEntry{
				category: category,
				txType:   rule.Type,
			}
			switch strings.ToLower(rule.Type) {
			case "debit":
				c.debitKeywords[key] = entry
			case "credit":
				c.creditKeywords[key] = entry
			default:
				log.Printf("[RulesCategorizer] категорія %q має невідомий тип %q — пропускаємо ключове слово %q", category, rule.Type, kw)
			}
		}
	}
}

// Categorize визначає категорію транзакції.
// Спочатку фільтрує правила за типом транзакції, потім шукає ключові слова.
func (c *RulesCategorizer) Categorize(tx *models.Transaction) string {
	desc := normalize(tx.Description + " " + tx.Counterparty)

	// Обираємо відповідний кеш залежно від типу транзакції
	var keywords map[string]keywordEntry
	switch tx.Type {
	case models.Credit:
		keywords = c.creditKeywords
	default: // models.Debit
		keywords = c.debitKeywords
	}

	for kw, entry := range keywords {
		if strings.Contains(desc, kw) {
			return entry.category
		}
	}

	return "Uncategorized"
}

// CategorizeAll категоризує всі транзакції у списку.
// Пропускає транзакції що вже мають категорію.
func (c *RulesCategorizer) CategorizeAll(txs []*models.Transaction) {
	for _, tx := range txs {
		if tx.Category == "" {
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
		log.Printf("[RulesCategorizer] AddRule: невідомий тип %q для категорії %q — правило не додано", txType, category)
		return
	}

	// Оновлюємо categories для подальшого збереження в YAML
	rule := c.categories[category]
	rule.Keywords = append(rule.Keywords, keyword)
	if rule.Type == "" && txType != "" {
		rule.Type = txType
	}
	c.categories[category] = rule
}

// normalize приводить рядок до єдиного вигляду:
// NFKD декомпозиція + видалення діакритики + lowercase.
// Вирішує проблему омогліфів (латинська i vs українська і).
func normalize(s string) string {
	t := transform.Chain(
		norm.NFKD,
		runes.Remove(runes.In(unicode.Mn)),
	)
	result, _, _ := transform.String(t, strings.ToLower(s))
	return result
}
