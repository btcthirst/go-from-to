// Package reader надає функціональність для читання файлів Excel.
package reader

import (
	"excel-parser/internal/model"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// GetAccrualRecords читає файл нарахування та повертає список записів з ПІБ та рахунками
func GetAccrualRecords(filename string) ([]model.AccrualRecord, error) {
	// Визначаємо тип файлу за розширенням
	var rows [][]string

	if strings.HasSuffix(filename, ".ods") {
		odsReader, err := OpenODS(filename)
		if err != nil {
			return nil, fmt.Errorf("помилка відкриття ODS файлу: %w", err)
		}

		var errGetSheet error
		rows, errGetSheet = odsReader.GetSheet("")
		if errGetSheet != nil {
			return nil, fmt.Errorf("помилка читання аркуша ODS: %w", errGetSheet)
		}
	} else {
		// XLSX
		f, err := excelize.OpenFile(filename)
		if err != nil {
			return nil, fmt.Errorf("помилка відкриття файлу нарахування: %w", err)
		}
		defer f.Close()

		sheetName := f.GetSheetName(f.GetActiveSheetIndex())
		excelRows, err := f.Rows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("помилка читання рядків: %w", err)
		}

		for excelRows.Next() {
			cols, err := excelRows.Columns()
			if err != nil {
				continue
			}
			rows = append(rows, cols)
		}
	}

	var records []model.AccrualRecord
	lineNum := 0

	// Пропускаємо заголовок (перших 6 рядків)
	for lineNum < 6 && lineNum < len(rows) {
		lineNum++
	}

	// Читаємо дані з рядка 7+
	for lineNum < len(rows) {
		cols := rows[lineNum]
		lineNum++

		// Пропускаємо порожні рядки
		if len(cols) == 0 {
			continue
		}

		// Очікуємо мінімум 3 колонки: № квартири, ПІБ та рахунок
		if len(cols) < 3 {
			continue
		}

		fullName := strings.TrimSpace(cols[1])
		account := strings.TrimSpace(cols[2])

		if fullName == "" || account == "" {
			continue
		}

		records = append(records, model.AccrualRecord{
			FullName: fullName,
			Account:  account,
		})
	}

	return records, nil
}

// GetPaymentRecords читає файл виписки та повертає список платежів
func GetPaymentRecords(filename string) ([]model.PaymentRecord, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("помилка відкриття файлу виписки: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("помилка читання рядків: %w", err)
	}

	var records []model.PaymentRecord
	lineNum := 0
	started := false

	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			lineNum++
			continue
		}

		lineNum++

		// Шукаємо заголовок таблиці платежів (містить "Дата проводки")
		if len(cols) > 0 && strings.Contains(cols[0], "Дата") {
			started = true
			continue
		}

		if !started {
			continue
		}

		// Пропускаємо порожні рядки
		if len(cols) == 0 {
			continue
		}

		// Очікуємо мінімум 9 колонок
		if len(cols) < 9 {
			continue
		}

		date := strings.TrimSpace(cols[1])       // Дата проводки
		sumStr := strings.TrimSpace(cols[3])     // Сума
		purpose := strings.TrimSpace(cols[5])    // Призначення платежу
		counterpartyName := strings.TrimSpace(cols[7])   // Назва контрагента
		counterpartyAccount := strings.TrimSpace(cols[8]) // Рахунок контрагента

		if date == "" || sumStr == "" {
			continue
		}

		// Конвертуємо суму у float64
		var sum float64
		_, err = fmt.Sscanf(sumStr, "%f", &sum)
		if err != nil {
			continue
		}

		// Комбінуємо дані для пошуку контрагента
		counterparty := counterpartyName
		if counterpartyAccount != "" {
			counterparty = counterpartyName + " (" + counterpartyAccount + ")"
		}
		if purpose != "" && counterparty == "" {
			counterparty = purpose
		}

		records = append(records, model.PaymentRecord{
			Date:         date,
			Sum:          sum,
			Counterparty: counterparty,
			OriginalData: counterpartyAccount,
		})
	}

	return records, nil
}

// FindAccountInCounterparty шукає рахунок у даних контрагента
func FindAccountInCounterparty(counterparty string, accounts []string) string {
	for _, account := range accounts {
		if strings.Contains(counterparty, account) {
			return account
		}
	}
	return ""
}

// trimRow видаляє пусті елементи з рядка
func trimRow(row []string) []string {
	result := []string{}
	for _, r := range row {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
