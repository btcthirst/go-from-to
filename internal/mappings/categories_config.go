package mappings

import (
	"log"
	"os"

	"bank-analyzer/assets"

	"gopkg.in/yaml.v3"
)

// CategoryRule — правило категоризації для однієї категорії.
type CategoryRule struct {
	Type     string   `yaml:"type"`
	Keywords []string `yaml:"keywords"`
}

// CategoriesConfig — мапа: назва категорії → правило.
type CategoriesConfig map[string]CategoryRule

func (c CategoriesConfig) Names() []string {
	names := make([]string, 0, len(c))
	for name := range c {
		names = append(names, name)
	}
	return names
}

func (c CategoriesConfig) KeywordsFor(category string) []string {
	rule, ok := c[category]
	if !ok {
		return nil
	}
	out := make([]string, len(rule.Keywords))
	copy(out, rule.Keywords)
	return out
}

// LoadCategoriesConfig завантажує конфігурацію категорій з YAML файлу.
// При помилці — повертає вбудований файл (застосунок не падає).
func LoadCategoriesConfig(path string) CategoriesConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[categories] warn: не вдалося прочитати %s: %v, використовую вбудований файл", path, err)
		return mustParseCategoriesConfig(assets.CategoriesYAML)
	}

	cfg, err := parseCategoriesConfig(data)
	if err != nil {
		log.Printf("[categories] warn: не вдалося розпарсити %s: %v, використовую вбудований файл", path, err)
		return mustParseCategoriesConfig(assets.CategoriesYAML)
	}

	if len(cfg) == 0 {
		log.Printf("[categories] warn: %s порожній, використовую вбудований файл", path)
		return mustParseCategoriesConfig(assets.CategoriesYAML)
	}

	return cfg
}

func parseCategoriesConfig(data []byte) (CategoriesConfig, error) {
	var cfg CategoriesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func mustParseCategoriesConfig(data []byte) CategoriesConfig {
	cfg, err := parseCategoriesConfig(data)
	if err != nil {
		panic("вбудований categories.yaml пошкоджений: " + err.Error())
	}
	return cfg
}
