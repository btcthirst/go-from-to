// Package config provides configuration management for the application.
package config

import "bank-analyzer/internal/mappings"

// Config зберігає налаштування застосунку та категорії транзакцій.
type Config struct {
	theme         string
	ConfigPath    string
	MappingsPath  string
	ReportPath    string
	ProvidersPath string
	Mappings      mappings.MappingsConfig
	Categories    mappings.CategoriesConfig
	Report311     mappings.Report311Config
	Providers     mappings.ProvidersConfig
}

func (c *Config) GetTheme() string { return c.theme }

func Load() *Config {
	const (
		categoriesPath = "assets/categories.yaml"
		mappingsPath   = "assets/column_mappings.yaml"
		report311Path  = "assets/report_311.yaml"
		providersPath  = "assets/providers.yaml"
	)

	return &Config{
		theme:         "system",
		ConfigPath:    categoriesPath,
		MappingsPath:  mappingsPath,
		ReportPath:    report311Path,
		ProvidersPath: providersPath,
		Mappings:      mappings.Load(mappingsPath),
		Categories:    mappings.LoadCategoriesConfig(categoriesPath),
		Report311:     mappings.LoadReport311Config(report311Path),
		Providers:     mappings.LoadProvidersConfig(providersPath),
	}
}
