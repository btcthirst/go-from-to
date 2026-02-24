package config

import "strings"

// CategoryNames повертає список назв усіх категорій.
func (c *Config) CategoryNames() []string {
	return c.categoryNames
}

// KeywordsForCategory повертає ключові слова для конкретної категорії.
func (c *Config) KeywordsForCategory(category string) []string {
	keywords, ok := c.keywords[category]
	if !ok {
		return nil
	}
	// Повертаємо копію щоб уникнути зовнішньої мутації
	result := make([]string, len(keywords))
	copy(result, keywords)
	return result
}

// AddCategory додає нову порожню категорію.
func (c *Config) AddCategory(name string) {
	// Уникаємо дублікатів
	for _, existing := range c.categoryNames {
		if strings.EqualFold(existing, name) {
			return
		}
	}
	c.categoryNames = append(c.categoryNames, name)
	if c.keywords == nil {
		c.keywords = make(map[string][]string)
	}
	c.keywords[name] = []string{}
}

// RemoveCategory видаляє категорію та її ключові слова.
func (c *Config) RemoveCategory(name string) {
	for i, cat := range c.categoryNames {
		if cat == name {
			c.categoryNames = append(c.categoryNames[:i], c.categoryNames[i+1:]...)
			break
		}
	}
	delete(c.keywords, name)
}

// AddKeyword додає ключове слово до категорії.
func (c *Config) AddKeyword(category, keyword string) {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return
	}
	// Уникаємо дублікатів
	for _, existing := range c.keywords[category] {
		if existing == kw {
			return
		}
	}
	c.keywords[category] = append(c.keywords[category], kw)
}

// RemoveKeyword видаляє ключове слово з категорії.
func (c *Config) RemoveKeyword(category, keyword string) {
	kws := c.keywords[category]
	for i, kw := range kws {
		if kw == keyword {
			c.keywords[category] = append(kws[:i], kws[i+1:]...)
			return
		}
	}
}
