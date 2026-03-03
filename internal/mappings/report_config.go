package mappings

import (
	"fmt"
	"log"
	"os"

	"bank-analyzer/assets"

	"gopkg.in/yaml.v3"
)

// Report311Config — конфігурація звіту "Журнал-ордер 311".
type Report311Config struct {
	MainCategories []string       `yaml:"main_categories"`
	SubColumns     []Report311Col `yaml:"sub_columns"`
}

// Report311Col — один стовпець групи "С Кт 311 в Дт рахунків".
type Report311Col struct {
	Account  string `yaml:"account"`
	Category string `yaml:"category"`
}

// LoadReport311Config завантажує конфігурацію звіту з YAML файлу.
// При помилці — повертає вбудований файл.
func LoadReport311Config(path string) Report311Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[report311] warn: не вдалося прочитати %s: %v, використовую вбудований файл", path, err)
		return mustParseReport311Config(assets.Report311YAML)
	}

	cfg, err := parseReport311Config(data)
	if err != nil {
		log.Printf("[report311] warn: не вдалося розпарсити %s: %v, використовую вбудований файл", path, err)
		return mustParseReport311Config(assets.Report311YAML)
	}

	return cfg
}

func parseReport311Config(data []byte) (Report311Config, error) {
	var cfg Report311Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Report311Config{}, err
	}
	return cfg, nil
}

func mustParseReport311Config(data []byte) Report311Config {
	cfg, err := parseReport311Config(data)
	if err != nil {
		panic("вбудований report_311.yaml пошкоджений: " + err.Error())
	}
	return cfg
}

// ValidateReport311Config перевіряє що всі категорії зі звіту 311
// присутні у конфігурації категорій.
// Повертає список попереджень — порожній якщо все коректно.
func ValidateReport311Config(cfg Report311Config, cats CategoriesConfig) []string {
	var warnings []string

	for _, sc := range cfg.SubColumns {
		if _, ok := cats[sc.Category]; !ok {
			warnings = append(warnings,
				fmt.Sprintf("report_311: категорія %q (рах. %s) не знайдена у categories.yaml",
					sc.Category, sc.Account),
			)
		}
	}

	for _, mc := range cfg.MainCategories {
		if _, ok := cats[mc]; !ok {
			warnings = append(warnings,
				fmt.Sprintf("report_311: основна категорія %q не знайдена у categories.yaml",
					mc),
			)
		}
	}

	return warnings
}
