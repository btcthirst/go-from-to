// Package mappings завантажує та надає маппінги колонок для різних форматів виписок.
package mappings

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// ParserMapping — маппінг колонок для одного парсера.
// Ключ — логічне поле (date, amount, ...), значення — список можливих назв колонок.
type ParserMapping map[string][]string

// MappingsConfig — маппінги для всіх парсерів.
type MappingsConfig map[string]ParserMapping

// Get повертає список кандидатів для конкретного поля.
// Якщо поле не знайдено — повертає порожній список (парсер обробить як -1).
func (m ParserMapping) Get(field string) []string {
	if candidates, ok := m[field]; ok {
		return candidates
	}
	return nil
}

var defaultMappings = MappingsConfig{
	"privatbank_csv": {
		"date":         {"Дата операції", "Дата проводки", "Дата"},
		"amount":       {"Сума", "Оборот", "Сума операції"},
		"currency":     {"Валюта"},
		"description":  {"Призначення платежу", "Призначення", "Опис"},
		"counterparty": {"Кореспондент", "Контрагент", "Найменування контрагента"},
		"iban":         {"Рахунок кореспондента", "IBAN кореспондента", "Рахунок контрагента"},
		"edrpou":       {"ЄДРПОУ кореспондента", "ЄДРПОУ контрагента"},
		"doc_num":      {"Номер документу", "№ документу", "Документ"},
	},
	"privatbank_xlsx": {
		"date":         {"Дата проводки", "Дата і час проводки", "Дата операції", "Дата"},
		"amount":       {"Сума в валюті рахунку", "Сума операції", "Сума", "Оборот"},
		"currency":     {"Валюта рахунку", "Валюта", "Валюта операції"},
		"description":  {"Призначення платежу", "Призначення", "Деталі операції", "Опис"},
		"counterparty": {"Назва контрагента", "Контрагент", "Найменування контрагента", "Отримувач/Платник"},
		"iban":         {"Рахунок контрагента", "IBAN контрагента", "Рахунок кореспондента", "Рахунок", "IBAN"},
		"balance":      {"Залишок після проводки", "Залишок", "Баланс"},
		"debit":        {"Дебет", "Витрата", "Сума дебету"},
		"credit":       {"Кредит", "Надходження", "Сума кредиту"},
	},
}

// Load завантажує маппінги з YAML файлу.
// Якщо файл не знайдено або пошкоджений — повертає вбудовані дефолтні маппінги
// і логує попередження (застосунок не падає).
func Load(path string) MappingsConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		// Файл відсутній — використовуємо вбудовані дефолти
		return defaultMappings
	}

	var cfg MappingsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Пошкоджений YAML — використовуємо вбудовані дефолти
		return defaultMappings
	}

	// Merge: yaml перезаписує дефолти тільки для тих парсерів що є у файлі.
	// Парсери яких немає у yaml — залишаються з дефолтів.
	merged := make(MappingsConfig, len(defaultMappings))
	for k, v := range defaultMappings {
		merged[k] = v
	}
	for k, v := range cfg {
		merged[k] = v
	}

	return merged
}

// ForParser повертає маппінг для конкретного парсера.
// Якщо парсер не знайдено — повертає помилку.
func (c MappingsConfig) ForParser(parserKey string) (ParserMapping, error) {
	m, ok := c[parserKey]
	if !ok {
		return nil, fmt.Errorf("маппінг для парсера %q не знайдено", parserKey)
	}
	return m, nil
}
