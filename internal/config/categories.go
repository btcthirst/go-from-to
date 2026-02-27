package config

import (
	"bank-analyzer/internal/mappings"
	"strings"
)

// CategoryNames повертає список назв усіх категорій.
func (c *Config) CategoryNames() []string {
	return c.Categories.Names()
}

// KeywordsForCategory повертає ключові слова для конкретної категорії.
func (c *Config) KeywordsForCategory(category string) []string {
	return c.Categories.KeywordsFor(category)
}

// AddCategory додає нову порожню категорію.
func (c *Config) AddCategory(name string) {
	for _, existing := range c.Categories.Names() {
		if strings.EqualFold(existing, name) {
			return
		}
	}
	if _, exists := c.Categories[name]; !exists {
		c.Categories[name] = mappings.CategoryRule{}
	} // ініціалізує порожнім значенням якщо нема
}

// RemoveCategory видаляє категорію та її ключові слова.
func (c *Config) RemoveCategory(name string) {
	delete(c.Categories, name)
}

// AddKeyword додає ключове слово до категорії.
func (c *Config) AddKeyword(category, keyword string) {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return
	}
	rule := c.Categories[category]
	for _, existing := range rule.Keywords {
		if existing == kw {
			return
		}
	}
	rule.Keywords = append(rule.Keywords, kw)
	c.Categories[category] = rule
}

// RemoveKeyword видаляє ключове слово з категорії.
func (c *Config) RemoveKeyword(category, keyword string) {
	rule := c.Categories[category]
	for i, kw := range rule.Keywords {
		if kw == keyword {
			rule.Keywords = append(rule.Keywords[:i], rule.Keywords[i+1:]...)
			c.Categories[category] = rule
			return
		}
	}
}
