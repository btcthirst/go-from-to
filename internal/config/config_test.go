package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestLoad тестує завантаження конфігурації
func TestLoad(t *testing.T) {
	// Створюємо тимчасовий файл з конфігурацією
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	testConfig := Config{
		Documents: map[string]DocumentTypeConfig{
			"test_doc": {
				Name:         "Test Document",
				DataStartRow: 5,
				HeaderEndRow: 4,
				MinColumns:   2,
				Columns: map[string]ColumnConfig{
					"name": {Index: 0, Name: "Name"},
					"id":   {Index: 1, Name: "ID"},
				},
				RequiredColumns: []string{"name", "id"},
			},
		},
	}

	// Записуємо конфіг у файл
	data, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Тестуємо завантаження
	err = Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if GlobalConfig == nil {
		t.Fatal("GlobalConfig is nil after Load()")
	}

	if _, exists := GlobalConfig.Documents["test_doc"]; !exists {
		t.Fatal("test_doc not found in GlobalConfig")
	}
}

// TestLoadNonExistentFile тестує завантаження неіснуючого файлу
func TestLoadNonExistentFile(t *testing.T) {
	err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Load() should return error for non-existent file")
	}

	if !strings.Contains(err.Error(), "помилка читання") {
		t.Errorf("expected 'помилка читання' in error, got: %v", err)
	}
}

// TestGetDocumentConfig тестує отримання конфігурації документа
func TestGetDocumentConfig(t *testing.T) {
	// Встановлюємо глобальну конфігурацію для тесту
	GlobalConfig = &Config{
		Documents: map[string]DocumentTypeConfig{
			"accruals": {
				Name:         "Accruals",
				DataStartRow: 7,
				HeaderEndRow: 6,
				MinColumns:   3,
			},
		},
	}

	config, err := GetDocumentConfig("accruals")
	if err != nil {
		t.Fatalf("GetDocumentConfig() failed: %v", err)
	}

	if config.Name != "Accruals" {
		t.Errorf("expected Name='Accruals', got %q", config.Name)
	}

	if config.DataStartRow != 7 {
		t.Errorf("expected DataStartRow=7, got %d", config.DataStartRow)
	}
}

// TestGetDocumentConfigNotFound тестує отримання неіснуючої конфігурації
func TestGetDocumentConfigNotFound(t *testing.T) {
	GlobalConfig = &Config{
		Documents: map[string]DocumentTypeConfig{
			"accruals": {Name: "Accruals"},
		},
	}

	_, err := GetDocumentConfig("nonexistent")
	if err == nil {
		t.Fatal("GetDocumentConfig() should return error for non-existent document")
	}

	if !strings.Contains(err.Error(), "не знайдена") {
		t.Errorf("expected 'не знайдена' in error, got: %v", err)
	}
}

// TestGetDocumentConfigNoConfig тестує отримання конфігурації без її завантаження
func TestGetDocumentConfigNoConfig(t *testing.T) {
	GlobalConfig = nil

	_, err := GetDocumentConfig("accruals")
	if err == nil {
		t.Fatal("GetDocumentConfig() should return error when GlobalConfig is nil")
	}

	if !strings.Contains(err.Error(), "не завантажена") {
		t.Errorf("expected 'не завантажена' in error, got: %v", err)
	}
}

// TestGetColumnIndex тестує отримання індекса колонки
func TestGetColumnIndex(t *testing.T) {
	dc := &DocumentTypeConfig{
		Columns: map[string]ColumnConfig{
			"name":    {Index: 0, Name: "Name"},
			"account": {Index: 2, Name: "Account"},
		},
	}

	tests := []struct {
		columnName    string
		expectedIndex int
		shouldError   bool
	}{
		{"name", 0, false},
		{"account", 2, false},
		{"nonexistent", -1, true},
	}

	for _, tt := range tests {
		index, err := dc.GetColumnIndex(tt.columnName)
		if tt.shouldError && err == nil {
			t.Errorf("GetColumnIndex(%q) should return error", tt.columnName)
		}
		if !tt.shouldError && err != nil {
			t.Errorf("GetColumnIndex(%q) failed: %v", tt.columnName, err)
		}
		if !tt.shouldError && index != tt.expectedIndex {
			t.Errorf("GetColumnIndex(%q) = %d, want %d", tt.columnName, index, tt.expectedIndex)
		}
	}
}

// TestIsValidRow тестує валідацію рядків за конфігурацією
func TestIsValidRow(t *testing.T) {
	dc := &DocumentTypeConfig{
		MinColumns: 3,
		Columns: map[string]ColumnConfig{
			"name":    {Index: 0},
			"account": {Index: 1},
		},
		RequiredColumns: []string{"name", "account"},
	}

	tests := []struct {
		name      string
		row       []string
		isValid   bool
	}{
		{
			name:    "Valid row with all required columns",
			row:     []string{"Іван Петренко", "1001", "extra"},
			isValid: true,
		},
		{
			name:    "Invalid row - missing required column",
			row:     []string{"Іван Петренко", "", "extra"},
			isValid: false,
		},
		{
			name:    "Invalid row - too few columns",
			row:     []string{"Іван Петренко"},
			isValid: false,
		},
		{
			name:    "Invalid row - empty",
			row:     []string{},
			isValid: false,
		},
		{
			name:    "Valid row - exactly min columns",
			row:     []string{"Name", "Account", "Extra"},
			isValid: true,
		},
	}

	for _, tt := range tests {
		result := dc.IsValidRow(tt.row)
		if result != tt.isValid {
			t.Errorf("IsValidRow(%v) = %v, want %v", tt.row, result, tt.isValid)
		}
	}
}

// BenchmarkGetDocumentConfig вимірює продуктивність отримання конфігурації
func BenchmarkGetDocumentConfig(b *testing.B) {
	GlobalConfig = &Config{
		Documents: map[string]DocumentTypeConfig{
			"accruals": {Name: "Accruals"},
			"payments": {Name: "Payments"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDocumentConfig("accruals")
	}
}

// BenchmarkIsValidRow вимірює продуктивність валідації рядків
func BenchmarkIsValidRow(b *testing.B) {
	dc := &DocumentTypeConfig{
		MinColumns: 3,
		Columns: map[string]ColumnConfig{
			"name":    {Index: 0},
			"account": {Index: 1},
		},
		RequiredColumns: []string{"name", "account"},
	}

	row := []string{"Іван Петренко", "1001", "extra"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dc.IsValidRow(row)
	}
}
