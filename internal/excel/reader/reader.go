// Package reader надає функціональність для читання файлів Excel.
package reader

import (
	"excel-parser/internal/config"
	"excel-parser/internal/model"
	"fmt"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// isValidName перевіряє чи рядок є валідним ПІБ (а не числом)
func isValidName(name string) bool {
	// Рядок повинен містити букви
	hasLetters := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') {
			hasLetters = true
			break
		}
	}

	if !hasLetters {
		return false
	}

	// Перевіряємо що це не чистий номер (з кількома символами як розділювачі)
	cleaned := regexp.MustCompile(`[0-9,.\s\-]`).ReplaceAllString(name, "")
	return len(cleaned) > 0
}

// GetAccrualRecords читає файл нарахування та повертає список записів з ПІБ та рахунками
func GetAccrualRecords(filename string) ([]model.AccrualRecord, error) {
	// Отримуємо конфігурацію для акордів
	accrualConfig, err := config.GetDocumentConfig("accruals")
	if err != nil {
		return nil, fmt.Errorf("помилка завантаження конфігурації: %w", err)
	}

	// Визначаємо тип файлу за розширенням
	var rows [][]string

	if strings.HasSuffix(filename, ".ods") {
		odsReader, err := OpenODS(filename)
		if err != nil {
			return nil, fmt.Errorf("помилка відкриття ODS файлу: %w", err)
		}

		// Отримуємо перший аркуш
		var errGetSheet error
		rows, errGetSheet = odsReader.GetSheet("")
		if errGetSheet != nil {
			return nil, fmt.Errorf("помилка читання аркуша ODS: %w", errGetSheet)
		}

		fmt.Printf("DEBUG: ODS файл завантажений, рядків: %d\n", len(rows))
		if len(rows) > 0 {
			fmt.Printf("DEBUG: Перший рядок: %v\n", rows[0])
			if len(rows) > accrualConfig.DataStartRow {
				fmt.Printf("DEBUG: Рядок %d (перший після заголовка): %v\n", accrualConfig.DataStartRow+1, rows[accrualConfig.DataStartRow])
			}
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

	// Пропускаємо заголовок згідно конфігурації
	headerEndRow := accrualConfig.HeaderEndRow
	for lineNum < headerEndRow && lineNum < len(rows) {
		fmt.Printf("DEBUG: Пропускаємо заголовок рядок %d: %v\n", lineNum+1, rows[lineNum])
		lineNum++
	}

	// Отримуємо індекси колонок з конфігурації
	fullNameIndex, _ := accrualConfig.GetColumnIndex("full_name")
	accountIndex, _ := accrualConfig.GetColumnIndex("account")

	// Читаємо дані
	for lineNum < len(rows) {
		cols := rows[lineNum]
		lineNum++

		if lineNum <= accrualConfig.DataStartRow+5 {
			fmt.Printf("DEBUG: Processing рядок %d: %v\n", lineNum, cols)
		}

		// Перевіряємо валідність рядка за конфігурацією
		if !accrualConfig.IsValidRow(cols) {
			continue
		}

		fullName := strings.TrimSpace(cols[fullNameIndex])
		account := strings.TrimSpace(cols[accountIndex])

		// Пропускаємо запису якщо ПІБ це не валідне ім'я
		if !isValidName(fullName) {
			if lineNum <= accrualConfig.DataStartRow+5 && (fullName != "" || account != "") {
				fmt.Printf("DEBUG: Skipped (invalid name): name='%s' account='%s'\n", fullName, account)
			}
			continue
		}

		fmt.Printf("DEBUG: Added record: name='%s' account='%s'\n", fullName, account)
		records = append(records, model.AccrualRecord{
			FullName: fullName,
			Account:  account,
		})
	}

	fmt.Printf("DEBUG: Total records found: %d\n", len(records))
	return records, nil
}

// GetPaymentRecords читає файл виписки та повертає список платежів
func GetPaymentRecords(filename string) ([]model.PaymentRecord, error) {
	// Отримуємо конфігурацію для платежів
	paymentConfig, err := config.GetDocumentConfig("payments")
	if err != nil {
		return nil, fmt.Errorf("помилка завантаження конфігурації: %w", err)
	}

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

	// Отримуємо індекси колонок з конфігурації
	dateIndex, _ := paymentConfig.GetColumnIndex("date")
	amountIndex, _ := paymentConfig.GetColumnIndex("amount")
	purposeIndex, _ := paymentConfig.GetColumnIndex("purpose")
	counterpartyNameIndex, _ := paymentConfig.GetColumnIndex("counterparty_name")
	counterpartyAccountIndex, _ := paymentConfig.GetColumnIndex("counterparty_account")

	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			lineNum++
			continue
		}

		lineNum++

		// Шукаємо заголовок таблиці платежів (тригер з конфігурації)
		if !started && len(cols) > 0 && strings.Contains(cols[0], paymentConfig.HeaderTrigger) {
			started = true
			fmt.Printf("DEBUG: Знайдено заголовок платежів на рядку %d\n", lineNum)
			continue
		}

		if !started {
			continue
		}

		// Перевіряємо валідність рядка за конфігурацією
		if !paymentConfig.IsValidRow(cols) {
			continue
		}

		date := strings.TrimSpace(cols[dateIndex])
		sumStr := strings.TrimSpace(cols[amountIndex])
		purpose := strings.TrimSpace(cols[purposeIndex])
		counterpartyName := strings.TrimSpace(cols[counterpartyNameIndex])
		counterpartyAccount := strings.TrimSpace(cols[counterpartyAccountIndex])

		// Конвертуємо суму у float64
		var sum float64
		_, err = fmt.Sscanf(sumStr, "%f", &sum)
		if err != nil {
			continue
		}

		// Комбінуємо дані для пошуку контрагента
		counterparty := counterpartyName
		if counterpartyName != "" && purpose != "" {
			counterparty = counterpartyName + " " + purpose
		}

		fmt.Printf("DEBUG: Payment record - date: %s, sum: %.2f, counterparty: %s\n", date, sum, counterparty)

		records = append(records, model.PaymentRecord{
			Date:         date,
			Sum:          sum,
			Purpose:      purpose,
			Counterparty: counterparty,
			OriginalData: counterpartyAccount,
		})
	}

	fmt.Printf("DEBUG: Total payment records found: %d\n", len(records))
	return records, nil
}

// FindAccountInCounterparty шукає рахунок у даних контрагента за допомогою карти для O(1) пошуку
// Замість лінійного пошуку O(n) через список, використовує итерацію по ключам карти
func FindAccountInCounterparty(counterparty string, accountToName map[string]string) string {
	// Проходимо по всім рахункам у карті та шукаємо перший, що міститься у контрагенті
	for account := range accountToName {
		if strings.Contains(counterparty, account) {
			return account
		}
	}
	return ""
}
