// Package utils містить допоміжні функції для парсерів.
package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func ParseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("порожня дата")
	}
	for _, layout := range []string{
		"02.01.2006", "02.01.2006 15:04:05", "02.01.2006 15:04",
		"2006-01-02", "2006-01-02 15:04:05", "2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("невідомий формат: %q", s)
}

func ParseDecimal(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Zero, nil
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00a0", "")
	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")
	switch {
	case hasComma && hasDot:
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	case hasComma:
		s = strings.ReplaceAll(s, ",", ".")
	}
	return decimal.NewFromString(s)
}

func FirstMatch(idx map[string]int, candidates []string) int {
	for _, name := range candidates {
		if i, ok := idx[name]; ok {
			return i
		}
	}
	return -1
}

func SafeGet(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// BuildHeaderIndex будує індекс «назва колонки → позиція» з рядка заголовків.
// Пробіли навколо назв автоматично обрізаються.
// Використовується CSV та XLSX парсерами.
func BuildHeaderIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}

// IsEmptyRow повертає true якщо всі клітинки рядка порожні або складаються лише з пробілів.
func IsEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
