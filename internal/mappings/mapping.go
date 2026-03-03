package mappings

import (
	"fmt"
	"log"
	"os"

	"bank-analyzer/assets"

	"gopkg.in/yaml.v3"
)

// ParserMapping — маппінг колонок для одного парсера.
type ParserMapping map[string][]string

// MappingsConfig — маппінги для всіх парсерів.
type MappingsConfig map[string]ParserMapping

func (m ParserMapping) Get(field string) []string {
	if candidates, ok := m[field]; ok {
		return candidates
	}
	return nil
}

// Load завантажує маппінги з YAML файлу.
// При помилці — повертає вбудований файл.
// Merge: вбудований файл є базою, yaml з диску перезаписує тільки ті
// парсери що в ньому є.
func Load(path string) MappingsConfig {
	embedded := mustParseMappingsConfig(assets.ColumnMappingsYAML)

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[mappings] warn: не вдалося прочитати %s: %v, використовую вбудований файл", path, err)
		return embedded
	}

	cfg, err := parseMappingsConfig(data)
	if err != nil {
		log.Printf("[mappings] warn: не вдалося розпарсити %s: %v, використовую вбудований файл", path, err)
		return embedded
	}

	merged := make(MappingsConfig, len(embedded))
	for k, v := range embedded {
		merged[k] = v
	}
	for k, v := range cfg {
		merged[k] = v
	}

	return merged
}

func (c MappingsConfig) ForParser(parserKey string) (ParserMapping, error) {
	m, ok := c[parserKey]
	if !ok {
		return nil, fmt.Errorf("маппінг для парсера %q не знайдено", parserKey)
	}
	return m, nil
}

func parseMappingsConfig(data []byte) (MappingsConfig, error) {
	var cfg MappingsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func mustParseMappingsConfig(data []byte) MappingsConfig {
	cfg, err := parseMappingsConfig(data)
	if err != nil {
		panic("вбудований column_mappings.yaml пошкоджений: " + err.Error())
	}
	return cfg
}
