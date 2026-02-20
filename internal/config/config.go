// Package config надає функціональність для читання та управління конфігурацією документів
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ColumnConfig описує конфігурацію колонки
type ColumnConfig struct {
	Index int    `yaml:"index"`
	Name  string `yaml:"name"`
}

// DocumentTypeConfig описує конфігурацію для типу документа
type DocumentTypeConfig struct {
	Name             string                   `yaml:"name"`
	SupportedFormats []string                 `yaml:"supported_formats"`
	DataStartRow     int                      `yaml:"data_start_row"`
	HeaderEndRow     int                      `yaml:"header_end_row"`
	SheetName        string                   `yaml:"sheet_name"`
	HeaderTrigger    string                   `yaml:"header_trigger"`
	Columns          map[string]ColumnConfig `yaml:"columns"`
	MinColumns       int                      `yaml:"min_columns"`
	RequiredColumns  []string                 `yaml:"required_columns"`
}

// Config описує загальну конфігурацію документів
type Config struct {
	Documents map[string]DocumentTypeConfig `yaml:"documents"`
}

// GlobalConfig зберігає глобальну конфігурацію
var GlobalConfig *Config

// Load завантажує конфігурацію зі YAML файлу
func Load(configPath string) error {
	// Якщо шлях відносний, шукаємо його відносно робочої папки або папки проекту
	if !filepath.IsAbs(configPath) {
		// Спробуємо найти у поточній директорії
		if _, err := os.Stat(configPath); err != nil {
			// Спробуємо знайти у папці проекту
			configPath = filepath.Join(".", configPath)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("помилка читання конфігураційного файлу: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("помилка парсингу YAML: %w", err)
	}

	GlobalConfig = config
	return nil
}

// GetDocumentConfig повертає конфігурацію для типу документа
func GetDocumentConfig(docType string) (*DocumentTypeConfig, error) {
	if GlobalConfig == nil {
		return nil, fmt.Errorf("конфігурація не завантажена. Викличте Load() спочатку")
	}

	docConfig, exists := GlobalConfig.Documents[docType]
	if !exists {
		return nil, fmt.Errorf("конфігурація для типу документа '%s' не знайдена", docType)
	}

	return &docConfig, nil
}

// GetColumnIndex повертає індекс колонки за її назвою
func (dc *DocumentTypeConfig) GetColumnIndex(columnName string) (int, error) {
	col, exists := dc.Columns[columnName]
	if !exists {
		return -1, fmt.Errorf("колонка '%s' не знайдена у конфігурації", columnName)
	}
	return col.Index, nil
}

// IsValidRow перевіряє чи рядок валідний за конфігурацією
func (dc *DocumentTypeConfig) IsValidRow(row []string) bool {
	if len(row) < dc.MinColumns {
		return false
	}

	// Перевіряємо обов'язкові колонки
	for _, requiredCol := range dc.RequiredColumns {
		index, err := dc.GetColumnIndex(requiredCol)
		if err != nil {
			continue
		}

		// Перевіряємо що індекс в межах рядка
		if index >= len(row) {
			return false
		}

		// Перевіряємо що значення не порожне
		if row[index] == "" {
			return false
		}
	}

	return true
}
