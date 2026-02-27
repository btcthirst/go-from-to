// Package mappings — завантаження конфігурації категорій транзакцій.
// Патерн ідентичний до mapping.go та report_cofig.go: читає YAML з assets/,
// при помилці повертає вбудовані дефолти і логує попередження.
package mappings

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// CategoryRule — правило категоризації для однієї категорії.
// Відповідає структурі YAML у assets/categories.yaml.
type CategoryRule struct {
	// Type визначає до яких транзакцій застосовується правило:
	//   "debit"  — тільки витрати
	//   "credit" — тільки надходження
	Type     string   `yaml:"type"`
	Keywords []string `yaml:"keywords"`
}

// CategoriesConfig — мапа: назва категорії → правило.
// Ключі збігаються з Category у models.Transaction.
type CategoriesConfig map[string]CategoryRule

// Names повертає список назв усіх категорій (порядок не гарантований).
func (c CategoriesConfig) Names() []string {
	names := make([]string, 0, len(c))
	for name := range c {
		names = append(names, name)
	}
	return names
}

// KeywordsFor повертає список ключових слів для категорії.
// Якщо категорія не знайдена — повертає nil.
func (c CategoriesConfig) KeywordsFor(category string) []string {
	rule, ok := c[category]
	if !ok {
		return nil
	}
	out := make([]string, len(rule.Keywords))
	copy(out, rule.Keywords)
	return out
}

// вбудовані дефолти — збігаються зі змістом assets/categories.yaml
var defaultCategoriesConfig = CategoriesConfig{
	"Комунальні": {
		Type:     "debit",
		Keywords: []string{"водоканал", "вувкг", "парки", "хоек", "тепло", "електро", "оплата"},
	},
	"Комісія банку": {
		Type:     "debit",
		Keywords: []string{"комісія", "комiсiя", "обслуговування", "банківські"},
	},
	"Зарплата": {
		Type:     "debit",
		Keywords: []string{"зарплата", "зп"},
	},
	"Військовий": {
		Type:     "debit",
		Keywords: []string{"військовий", "вiйськовий збiр"},
	},
	"ЄСВ": {
		Type:     "debit",
		Keywords: []string{"єсв"},
	},
	"ПДФО": {
		Type:     "debit",
		Keywords: []string{"пдфо"},
	},
	"Внески": {
		Type:     "credit",
		Keywords: []string{"внесок", "внески", "оплата", "плата", "кварт плата", "квартплата"},
	},
	"Контрагенти": {
		Type:     "credit",
		Keywords: []string{"київстар", "воля", "норма", "телесвіт"},
	},
}

// LoadCategoriesConfig завантажує конфігурацію категорій з YAML файлу.
// При помилці — повертає вбудовані дефолти (застосунок не падає).
func LoadCategoriesConfig(path string) CategoriesConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[categories] warn: не вдалося прочитати %s: %v, використовую дефолти", path, err)
		return defaultCategoriesConfig
	}

	var cfg CategoriesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("[categories] warn: не вдалося розпарсити %s: %v, використовую дефолти", path, err)
		return defaultCategoriesConfig
	}

	if len(cfg) == 0 {
		log.Printf("[categories] warn: %s порожній, використовую дефолти", path)
		return defaultCategoriesConfig
	}

	return cfg
}
