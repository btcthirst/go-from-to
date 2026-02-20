// Package app - це основний пакет застосунку, який відповідає за ініціалізацію та виконання логіки обробки файлів Excel.
package app

import (
	"excel-parser/internal/excel/reader"
	"excel-parser/internal/model"
	"fmt"
	"log"
)

// ProcessPayments обробляє платежі та повертає результати (для UI)
func ProcessPayments(accrualFile, paymentFile string) ([]model.ResultRecord, error) {
	log.Println("Початок обробки документів...")

	// Крок 1: Читаємо дані з файлу нарахування
	log.Printf("Читання даних з файлу нарахування: %s\n", accrualFile)
	accruals, err := reader.GetAccrualRecords(accrualFile)
	if err != nil {
		return nil, err
	}
	log.Printf("Знайдено %d записів у файлі нарахування\n", len(accruals))

	// Крок 2: Створюємо карту для швидкого пошуку ПІБ по рахунку
	accountToName := make(map[string]string)
	for _, accrual := range accruals {
		accountToName[accrual.Account] = accrual.FullName
		log.Printf("  - ПІБ: %s, Рахунок: %s\n", accrual.FullName, accrual.Account)
	}

	// Крок 3: Читаємо дані з файлу виписки
	log.Printf("Читання даних з файлу виписки: %s\n", paymentFile)
	payments, err := reader.GetPaymentRecords(paymentFile)
	if err != nil {
		return nil, err
	}
	log.Printf("Знайдено %d платежів у файлі виписки\n", len(payments))

	// Крок 4: Обробляємо платежі та замінюємо рахунки на ПІБ
	results := processPayments(payments, accountToName)

	return results, nil
}

// GetAccruals читає та повертає записи про нарахування (для UI)
func GetAccruals(accrualFile string) ([]model.AccrualRecord, error) {
	log.Printf("Читання даних з файлу нарахування: %s\n", accrualFile)
	accruals, err := reader.GetAccrualRecords(accrualFile)
	if err != nil {
		return nil, err
	}
	log.Printf("Знайдено %d записів у файлі нарахування\n", len(accruals))
	return accruals, nil
}

// processPayments обробляє платежі, замінює рахунки на ПІБ та повертає результати
// Використовує карту accountToName для швидкого O(1) пошуку замість лінійного пошуку
func processPayments(payments []model.PaymentRecord, accountToName map[string]string) []model.ResultRecord {
	var results []model.ResultRecord

	for _, payment := range payments {
		// Визначаємо чи це відрахування (негативна сума)
		isDeduction := payment.Sum < 0

		result := model.ResultRecord{
			Date:         payment.Date,
			Sum:          payment.Sum,
			Counterparty: payment.Counterparty,
			IsDeduction:  isDeduction,
			Purpose:      payment.Purpose,
		}

		// Шукаємо рахунок у даних контрагента використовуючи карту для швидкого пошуку
		foundAccount := reader.FindAccountInCounterparty(payment.Counterparty, accountToName)
		if foundAccount != "" {
			// Знайшли рахунок - замінюємо контрагента на ПІБ
			fullName, exists := accountToName[foundAccount]
			if exists {
				result.Name = fullName
				result.Account = foundAccount
				result.Counterparty = fullName

				if isDeduction {
					fmt.Printf("Відрахування рахунок %s на користь %s, ПІБ: %s (%.2f)\n", foundAccount, payment.Purpose, fullName, payment.Sum)
				} else {
					fmt.Printf("Замінено рахунок %s на ПІБ: %s\n", foundAccount, fullName)
				}
			}
		}

		results = append(results, result)
	}

	return results
}
