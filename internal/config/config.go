// Package config provides configuration management for the application.
package config

import "bank-analyzer/internal/mappings"

// Config зберігає налаштування застосунку та категорії транзакцій.
type Config struct {
	theme         string
	ConfigPath    string
	MappingsPath  string
	Mappings      mappings.MappingsConfig
	categoryNames []string
	keywords      map[string][]string
}

func (c *Config) GetTheme() string { return c.theme }

func Load() *Config {
	mappingsPath := "assets/column_mappings.yaml"
	return &Config{
		theme:        "system",
		ConfigPath:   "assets/categories.yaml",
		MappingsPath: mappingsPath,
		Mappings:     mappings.Load(mappingsPath),
		categoryNames: []string{
			"Комунальні послуги", "Послуги банку", "Зарплата",
			"Податки", "Інші витрати", "Внески",
		},
		keywords: map[string][]string{
			"Комунальні послуги": {"водоканал", "тепло", "електро"},
			"Послуги банку":      {"комісія", "обслуговування", "банківські послуги"},
			"Зарплата":           {"зарплата", "зп", "виплата"},
			"Податки":            {"податок", "пдв", "військовий збір", "пдфо", "єсв"},
			"Інші витрати":       {"оплата за матеріал", "оплата за виконані роботи"},
			"Внески":             {"внесок", "оплата", "плата", "квартплата"},
		},
	}
}
