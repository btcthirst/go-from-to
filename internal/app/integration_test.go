package app

import (
	"excel-parser/internal/config"
	"excel-parser/internal/model"
	"os"
	"path/filepath"
	"testing"
)

// TestIntegration_FullWorkflow тестує повний робочий цикл обробки даних
func TestIntegration_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Створюємо тимчасовий каталог для конфіг
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	// Створюємо простий конфіг
	configContent := `
documents:
  accruals:
    name: "Test Accruals"
    data_start_row: 1
    header_end_row: 0
    columns:
      full_name:
        index: 0
        name: "ПІБ"
      account:
        index: 1
        name: "Рахунок"
    min_columns: 2
    required_columns:
      - full_name
      - account
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Завантажуємо конфіг
	if err := config.Load(configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Тестуємо обробку платежів
	payments := []model.PaymentRecord{
		{
			Date:         "2024-01-15",
			Sum:          100.00,
			Purpose:      "Комунальні",
			Counterparty: "ПАТ БТА BANCO 1001",
			OriginalData: "1001",
		},
		{
			Date:         "2024-01-16",
			Sum:          -50.00,
			Purpose:      "Податок",
			Counterparty: "ПОДАТКОВА СЛУЖБА 1002",
			OriginalData: "1002",
		},
	}

	accountToName := map[string]string{
		"1001": "Іван Петренко",
		"1002": "Марія Сидоренко",
	}

	results := processPayments(payments, accountToName)

	// Перевіряємо результати
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Перший платіж
	if results[0].Name != "Іван Петренко" {
		t.Errorf("payment 1: expected Name='Іван Петренко', got %q", results[0].Name)
	}

	// Другий платіж (відрахування)
	if !results[1].IsDeduction {
		t.Errorf("payment 2: expected IsDeduction=true")
	}
	if results[1].Name != "Марія Сидоренко" {
		t.Errorf("payment 2: expected Name='Марія Сидоренко', got %q", results[1].Name)
	}
}

// TestIntegration_LargePaymentBatch тестує обробку великої кількості платежів
func TestIntegration_LargePaymentBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Генеруємо велику кількість платежів
	payments := make([]model.PaymentRecord, 10000)
	for i := 0; i < 10000; i++ {
		accountNum := "1001"
		if i%2 == 0 {
			accountNum = "1002"
		}
		sum := 100.0 + float64(i%1000)
		if i%5 == 0 {
			sum = -sum // Кожен п'ятий платіж - відрахування
		}

		payments[i] = model.PaymentRecord{
			Date:         "2024-01-15",
			Sum:          sum,
			Purpose:      "Test payment",
			Counterparty: "ПАТ БТА BANCO платіж " + accountNum,
			OriginalData: accountNum,
		}
	}

	accountToName := map[string]string{
		"1001": "Іван Петренко",
		"1002": "Марія Сидоренко",
	}

	results := processPayments(payments, accountToName)

	// Перевіряємо кількість результатів
	if len(results) != 10000 {
		t.Errorf("expected 10000 results, got %d", len(results))
	}

	// Перевіряємо, що більшість платежів мають знайде рахунки
	foundCount := 0
	for _, r := range results {
		if r.Account != "" {
			foundCount++
		}
	}

	if foundCount < 9000 { // Мінімум 90% повинно мати рахунки
		t.Errorf("expected at least 9000 found accounts, got %d", foundCount)
	}
}

// TestIntegration_DeductionClassification тестує класифікацію платежів та відрахувань
func TestIntegration_DeductionClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	payments := []model.PaymentRecord{
		{Date: "2024-01-01", Sum: 100.00, Counterparty: "Контрагент 1001"},
		{Date: "2024-01-02", Sum: -50.00, Counterparty: "Контрагент 1001"},
		{Date: "2024-01-03", Sum: 200.50, Counterparty: "Контрагент 1001"},
		{Date: "2024-01-04", Sum: -100.25, Counterparty: "Контрагент 1001"},
		{Date: "2024-01-05", Sum: 0.00, Counterparty: "Контрагент 1001"},
	}

	accountToName := map[string]string{
		"1001": "Test User",
	}

	results := processPayments(payments, accountToName)

	// Перевіряємо класифікацію
	for i, result := range results {
		expectedIsDeduction := result.Sum < 0

		if result.IsDeduction != expectedIsDeduction {
			t.Errorf("payment %d: IsDeduction=%v, expected=%v for Sum=%v",
				i, result.IsDeduction, expectedIsDeduction, result.Sum)
		}
	}
}
