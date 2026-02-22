package categoryzer

import (
	"bank-analyzer/internal/models"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Rule struct {
	Keywords []string `yaml:"keywords"`
	MinAmt   float64  `yaml:"min_amount,omitempty"`
	MaxAmt   float64  `yaml:"max_amount,omitempty"`
	Type     string   `yaml:"type,omitempty"`
}

type CategoryConfig map[string]Rule

type RulesCategorizer struct {
	categories CategoryConfig
	keywordMap map[string]string
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
		categories: cfg,
		keywordMap: make(map[string]string),
	}

	for category, rule := range cfg {
		for _, keyword := range rule.Keywords {
			c.keywordMap[strings.ToLower(keyword)] = category
		}
	}
	return c, nil
}

func (c *RulesCategorizer) Categorize(tx *models.Transaction) string {
	desc := strings.ToLower(tx.Description + "  " + tx.Counterparty)
	for keyword, category := range c.keywordMap {
		if strings.Contains(desc, keyword) {
			return category
		}
	}
	return "Uncategorized"
}

func (c *RulesCategorizer) CategorizeAll(txs []*models.Transaction) {
	for _, tx := range txs {
		if tx.Category == "" {
			tx.Category = c.Categorize(tx)
		}
	}
}

func (c *RulesCategorizer) AddRule(category, keyword string) {
	kw := strings.ToLower(keyword)
	c.keywordMap[kw] = category

	rule := c.categories[category]
	rule.Keywords = append(rule.Keywords, kw)
	c.categories[category] = rule
}
