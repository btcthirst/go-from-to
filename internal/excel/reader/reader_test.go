package reader

import (
	"strings"
	"testing"
)

// TestFindAccountInCounterparty тестує пошук рахунку у даних контрагента
func TestFindAccountInCounterparty(t *testing.T) {
	tests := []struct {
		name            string
		counterparty    string
		accountToName   map[string]string
		expectedAccount string
	}{
		{
			name:         "Simple account found",
			counterparty: "ПАТ БТА BANCO платіж 1001",
			accountToName: map[string]string{
				"1001": "Іван Петренко",
				"1002": "Марія Сидоренко",
			},
			expectedAccount: "1001",
		},
		{
			name:         "Account not found",
			counterparty: "ПАТ БТА BANCO платіж 9999",
			accountToName: map[string]string{
				"1001": "Іван Петренко",
				"1002": "Марія Сидоренко",
			},
			expectedAccount: "",
		},
		{
			name:         "Empty counterparty",
			counterparty: "",
			accountToName: map[string]string{
				"1001": "Іван Петренко",
			},
			expectedAccount: "",
		},
		{
			name:            "Empty map",
			counterparty:    "ПАТ БТА BANCO платіж 1001",
			accountToName:   map[string]string{},
			expectedAccount: "",
		},
		{
			name:         "Multiple matching accounts (returns first found)",
			counterparty: "ПАТ БТА BANCO платіж 1001 1002",
			accountToName: map[string]string{
				"1001": "Іван Петренко",
				"1002": "Марія Сидоренко",
			},
			expectedAccount: "", // map iteration order is random, so we can't predict which one will be found first
		},
		{
			name:         "Account with spaces",
			counterparty: "ПАТ БТА BANCO платіж 1001-UA",
			accountToName: map[string]string{
				"1001-UA": "Іван Петренко",
			},
			expectedAccount: "1001-UA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindAccountInCounterparty(tt.counterparty, tt.accountToName)
			
			// Для тесту з декількома рахунками просто провіримо що щось знайдено
			if tt.name == "Multiple matching accounts (returns first found)" {
				if result == "" {
					t.Errorf("expected some account to be found, got empty string")
				}
				// Нічого більше - map iteration order непередбачуваний
				return
			}
			
			if tt.expectedAccount != "" && result != tt.expectedAccount {
				t.Errorf("expected %q, got %q", tt.expectedAccount, result)
			}
			if tt.expectedAccount == "" && result != "" {
				t.Errorf("expected empty string, got %q", result)
			}
		})
	}
}

// TestIsValidName тестує валідацію імен
func TestIsValidName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  bool
		testName  string
	}{
		{"Valid Ukrainian name", "Іван Петренко", true, "Ukrainian letters"},
		{"Valid Latin name", "John Smith", true, "Latin letters"},
		{"Valid mixed", "Іван Smith", true, "Mixed Latin and Cyrillic"},
		{"Number only", "12345", false, "Pure numbers"},
		{"Number with comma", "123,456", false, "Numbers with separator"},
		{"Empty string", "", false, "Empty string"},
		{"Whitespace only", "   ", false, "Whitespace"},
		{"Valid with numbers", "John 2 Smith", true, "Name with number in middle"},
		{"Russian letters", "Иван Петренко", true, "Russian letters"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := isValidName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// BenchmarkFindAccountInCounterparty вимірює продуктивність пошуку рахунків
func BenchmarkFindAccountInCounterparty(b *testing.B) {
	// Створюємо велику карту рахунків
	accountToName := make(map[string]string)
	for i := 1; i <= 10000; i++ {
		accNum := strings.Repeat("0", 5-len(string(rune(i)))) + string(rune(i))
		if i < 10 {
			accNum = "1000" + string(rune('0'+i))
		} else if i < 100 {
			accNum = "100" + string(rune('0'+(i/10))) + string(rune('0'+(i%10)))
		}
		accountToName[accNum] = "Test User " + string(rune(i))
	}

	counterparty := "ПАТ БТА BANCO платіж 5000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindAccountInCounterparty(counterparty, accountToName)
	}
}

// BenchmarkIsValidName вимірює продуктивність валідації імені
func BenchmarkIsValidName(b *testing.B) {
	testName := "Іван Петренко"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidName(testName)
	}
}
