// Package config provides configuration management for the application.
package config

import (
	"bank-analyzer/internal/mappings"
	"log"
)

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
	Warnings      []string // warnings collected during configuration loading
}

func (c *Config) GetTheme() string { return c.theme }

func Load() *Config {
	const (
		categoriesPath = "assets/categories.yaml"
		mappingsPath   = "assets/column_mappings.yaml"
		report311Path  = "assets/report_311.yaml"
		providersPath  = "assets/providers.yaml"
	)

	cfg := &Config{
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

	cfg.Warnings = mappings.ValidateReport311Config(cfg.Report311, cfg.Categories)

	// Validation: categories in report 311 must exist in categories.yaml
	if warnings := mappings.ValidateReport311Config(cfg.Report311, cfg.Categories); len(warnings) > 0 {
		for _, w := range warnings {
			log.Printf("[config] warn: %s", w)
		}
	}

	return cfg
}
