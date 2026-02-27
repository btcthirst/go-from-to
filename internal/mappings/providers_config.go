// Package mappings — завантаження довідника постачальників для категорії "Внески".
// Патерн ідентичний до categories_config.go та report_cofig.go.
package mappings

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// ProvidersConfig — конфігурація довідника постачальників.
type ProvidersConfig struct {
	// Providers — мапа: код (підрядок з опису/контрагента) → назва постачальника.
	// Наприклад: "40160" → "Пупкін А.Ф"
	Providers map[string]string `yaml:"providers"`
}

var defaultProvidersConfig = ProvidersConfig{
	Providers: map[string]string{},
}

// LoadProvidersConfig завантажує довідник з YAML файлу.
// При помилці — повертає порожній довідник (застосунок не падає).
func LoadProvidersConfig(path string) ProvidersConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[providers] warn: не вдалося прочитати %s: %v, використовую порожній довідник", path, err)
		return defaultProvidersConfig
	}

	var cfg ProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("[providers] warn: не вдалося розпарсити %s: %v, використовую порожній довідник", path, err)
		return defaultProvidersConfig
	}

	if cfg.Providers == nil {
		cfg.Providers = map[string]string{}
	}

	return cfg
}
