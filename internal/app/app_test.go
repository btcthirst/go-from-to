package app

import (
	"excel-parser/internal/model"
	"testing"
)

// TestProcessPayments тестує основну функцію обробки платежів
func TestProcessPayments_Basic(t *testing.T) {
	// Передбачувані платежі
	payments := []model.PaymentRecord{
		{
			Date:         "2024-01-15",
			Sum:          100.00,
			Purpose:      "Комунальні послуги",
			Counterparty: "ПАТ БТА BANCO платіж 1001",
			OriginalData: "1001-UA",
		},
		{
			Date:         "2024-01-16",
			Sum:          -50.00, // Відрахування
			Purpose:      "Податок на доходи",
			Counterparty: "ДЕРЖАВНА ПОДАТКОВА СЛУЖБА 1002",
			OriginalData: "1002-UA",
		},
		{
			Date:         "2024-01-17",
			Sum:          75.50,
			Purpose:      "Комісія",
			Counterparty: "НЕВІДОМИЙ КОНТРАГЕНТ 9999",
			OriginalData: "9999-UA",
		},
	}

	// Карта рахунків
	accountToName := map[string]string{
		"1001": "Іван Петренко",
		"1002": "Марія Сидоренко",
	}

	// Виконуємо обробку
	results := processPayments(payments, accountToName)

	// Перевіряємо результати
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Перший платіж - повинен бути знайдено
	if results[0].Name != "Іван Петренко" {
		t.Errorf("expected Name='Іван Петренко', got %q", results[0].Name)
	}
	if results[0].Account != "1001" {
		t.Errorf("expected Account='1001', got %q", results[0].Account)
	}
	if results[0].IsDeduction != false {
		t.Errorf("expected IsDeduction=false, got %v", results[0].IsDeduction)
	}

	// Другий платіж - відрахування
	if results[1].IsDeduction != true {
		t.Errorf("expected IsDeduction=true for negative sum, got %v", results[1].IsDeduction)
	}
	if results[1].Sum != -50.00 {
		t.Errorf("expected Sum=-50.00, got %v", results[1].Sum)
	}

	// Третій платіж - рахунок не знайдено
	if results[2].Name != "" {
		t.Errorf("expected Name='', got %q", results[2].Name)
	}
	if results[2].Account != "" {
		t.Errorf("expected Account='', got %q", results[2].Account)
	}
}

// TestProcessPayments_EmptyPayments тестує обробку порожного списку платежів
func TestProcessPayments_EmptyPayments(t *testing.T) {
	payments := []model.PaymentRecord{}
	accountToName := map[string]string{
		"1001": "Іван Петренко",
	}

	results := processPayments(payments, accountToName)

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty payments, got %d", len(results))
	}
}

// TestProcessPayments_EmptyAccountMap тестує обробку порожної карти рахунків
func TestProcessPayments_EmptyAccountMap(t *testing.T) {
	payments := []model.PaymentRecord{
		{
			Date:         "2024-01-15",
			Sum:          100.00,
			Purpose:      "Платіж",
			Counterparty: "Контрагент 1001",
			OriginalData: "1001-UA",
		},
	}
	accountToName := map[string]string{}

	results := processPayments(payments, accountToName)

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Рахунок не повинен бути знайдено
	if results[0].Name != "" {
		t.Errorf("expected Name='', got %q", results[0].Name)
	}
}

// TestProcessPayments_ConsistentData тестує консистентність даних у результатах
func TestProcessPayments_ConsistentData(t *testing.T) {
	payments := []model.PaymentRecord{
		{
			Date:         "2024-01-15",
			Sum:          100.00,
			Purpose:      "Комунальні",
			Counterparty: "BANCO 1001",
			OriginalData: "1001-UA",
		},
	}

	accountToName := map[string]string{
		"1001": "Іван Петренко",
	}

	results := processPayments(payments, accountToName)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]

	// Дата та сума повинні зберігатися
	if result.Date != "2024-01-15" {
		t.Errorf("expected Date='2024-01-15', got %q", result.Date)
	}
	if result.Sum != 100.00 {
		t.Errorf("expected Sum=100.00, got %v", result.Sum)
	}

	// Призначення повинно зберігатися
	if result.Purpose != "Комунальні" {
		t.Errorf("expected Purpose='Комунальні', got %q", result.Purpose)
	}

	// IsDeduction повинна залежати від знаку суми
	if result.Sum > 0 && result.IsDeduction {
		t.Error("expected IsDeduction=false for positive sum")
	}
}

// TestProcessPayments_MultipleMatches тестує випадок, коли рахунок знайдено
func TestProcessPayments_AccountFound(t *testing.T) {
	payments := []model.PaymentRecord{
		{
			Date:         "2024-01-15",
			Sum:          100.00,
			Purpose:      "Комунальні",
			Counterparty: "ПАТ БТА BANCO счёт 1001 UA313052990000026006021101792",
			OriginalData: "1001-UA",
		},
	}

	accountToName := map[string]string{
		"1001": "Іван Петренко",
	}

	results := processPayments(payments, accountToName)

	if results[0].Name != "Іван Петренко" {
		t.Errorf("expected Name='Іван Петренко', got %q", results[0].Name)
	}
	if results[0].Counterparty != "Іван Петренко" {
		t.Errorf("expected Counterparty='Іван Петренко', got %q", results[0].Counterparty)
	}
}

// BenchmarkProcessPayments вимірює продуктивність обробки платежів
func BenchmarkProcessPayments(b *testing.B) {
	// Створюємо велику кількість платежів
	payments := make([]model.PaymentRecord, 1000)
	for i := 0; i < 1000; i++ {
		payments[i] = model.PaymentRecord{
			Date:         "2024-01-15",
			Sum:          100.00 + float64(i),
			Purpose:      "Платіж",
			Counterparty: "ПАТ БТА BANCO платіж 1001",
			OriginalData: "1001-UA",
		}
	}

	// Створюємо карту рахунків
	accountToName := make(map[string]string)
	for i := 1; i <= 500; i++ {
		account := "1" + string(rune(i))
		accountToName[account] = "User " + string(rune(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processPayments(payments, accountToName)
	}
}
