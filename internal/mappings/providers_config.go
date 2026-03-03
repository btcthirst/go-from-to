// Package mappings — завантаження довідника постачальників для категорії "Внески".
package mappings

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// ProvidersConfig — конфігурація довідника постачальників.
type ProvidersConfig struct {
	// Providers — мапа: код (підрядок з опису/контрагента) → назва постачальника.
	// Наприклад: "40160" → "Пупкін А.Ф"
	Providers map[string]string `yaml:"providers"`

	// Loaded — true якщо файл був знайдений і успішно завантажений.
	// Використовується UI для показу підказки при першому запуску.
	Loaded bool `yaml:"-"`
}

// LoadProvidersConfig завантажує довідник з YAML файлу.
// Файл опціональний — відсутність не є помилкою, але без нього
// категорія "Внески" не матиме розбивки по постачальниках у звіті.
func LoadProvidersConfig(path string) ProvidersConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf(`[providers] файл %s не знайдено.
Для повної функціональності створіть файл з довідником постачальників.
Формат:
  providers:
    "12345": "Іваненко І.І"
    "67890": "Петренко П.П"
Розташування: поруч з бінарником або у робочій директорії.`, path)
		return ProvidersConfig{Providers: map[string]string{}}
	}

	var cfg ProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("[providers] warn: не вдалося розпарсити %s: %v", path, err)
		return ProvidersConfig{Providers: map[string]string{}}
	}

	if cfg.Providers == nil {
		cfg.Providers = map[string]string{}
	}

	log.Printf("[providers] завантажено %d постачальників з %s", len(cfg.Providers), path)
	cfg.Loaded = true
	return cfg
}
