// Package mappings — завантаження конфігурації звіту 311.
// Патерн ідентичний до mappings.Load: читає YAML з assets/,
// при помилці повертає вбудовані дефолти і логує попередження.
package mappings

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Report311Config — конфігурація звіту "Журнал-ордер 311".
type Report311Config struct {
	MainCategories []string       `yaml:"main_categories"`
	SubColumns     []Report311Col `yaml:"sub_columns"`
}

// Report311Col — один стовпець групи "С Кт 311 в Дт рахунків".
type Report311Col struct {
	Account  string `yaml:"account"`  // заголовок колонки: "313", "63" тощо
	Category string `yaml:"category"` // назва категорії транзакції
}

var defaultReport311Config = Report311Config{
	MainCategories: []string{"Внески", "Контрагенти"},
	SubColumns: []Report311Col{
		{Account: "313", Category: "Зарплата"},
		{Account: "63", Category: "Комунальні"},
		{Account: "641", Category: "ПДФО"},
		{Account: "641.1", Category: "Військовий"},
		{Account: "651", Category: "ЄСВ"},
		{Account: "94", Category: "Комісія банку"},
	},
}

// LoadReport311Config завантажує конфігурацію звіту з YAML файлу.
// При помилці — повертає вбудовані дефолти (застосунок не падає).
func LoadReport311Config(path string) Report311Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[report311] warn: не вдалося прочитати %s: %v, використовую дефолти", path, err)
		return defaultReport311Config
	}

	var cfg Report311Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("[report311] warn: не вдалося розпарсити %s: %v, використовую дефолти", path, err)
		return defaultReport311Config
	}

	return cfg
}
