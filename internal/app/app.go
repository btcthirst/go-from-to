// Package app - це основний пакет застосунку, який відповідає за ініціалізацію та виконання логіки обробки файлів Excel.
package app

import (
	"excel-parser/internal/excel/reader"
	"excel-parser/internal/excel/writer"
	"excel-parser/internal/model"
	"fmt"
	"log"
)

// Init ініціалізує застосунок та запускає процес обробки даних
func Init() {
	log.Println("Початок обробки документів...")

	// Файли для обробки
	accrualFile := "нарахування26.ods"
	paymentFile := "stmts_37465042_UA313052990000026006021101792_1766677585832.xlsx"

	// Крок 1: Читаємо дані з файлу нарахування
	log.Printf("Читання даних з файлу нарахування: %s\n", accrualFile)
	accruals, err := reader.GetAccrualRecords(accrualFile)
	if err != nil {
		log.Fatalf("Помилка при читанні файлу нарахування: %v\n", err)
	}
	log.Printf("Знайдено %d записів у файлі нарахування\n", len(accruals))

	// Крок 2: Створюємо карту для швидкого пошуку ПІБ по рахунку
	accountToName := make(map[string]string)
	accountList := []string{}
	for _, accrual := range accruals {
		accountToName[accrual.Account] = accrual.FullName
		accountList = append(accountList, accrual.Account)
		log.Printf("  - ПІБ: %s, Рахунок: %s\n", accrual.FullName, accrual.Account)
	}

	// Крок 3: Читаємо дані з файлу виписки
	log.Printf("Читання даних з файлу виписки: %s\n", paymentFile)
	payments, err := reader.GetPaymentRecords(paymentFile)
	if err != nil {
		log.Fatalf("Помилка при читанні файлу виписки: %v\n", err)
	}
	log.Printf("Знайдено %d платежів у файлі виписки\n", len(payments))

	// Крок 4: Обробляємо платежі та замінюємо рахунки на ПІБ
	results := processPayments(payments, accountToName, accountList)

	// Крок 5: Записуємо результати
	log.Println("Запис результатів у файл summery.xlsx...")
	if err := writer.WriteResults(results); err != nil {
		log.Fatalf("Помилка при записі результатів: %v\n", err)
	}

	log.Println("Обробка завершена успішно!")
}

// processPayments обробляє платежі, замінює рахунки на ПІБ та повертає результати
func processPayments(payments []model.PaymentRecord, accountToName map[string]string, accountList []string) []model.ResultRecord {
	var results []model.ResultRecord

	for _, payment := range payments {
		result := model.ResultRecord{
			Date:         payment.Date,
			Sum:          payment.Sum,
			Counterparty: payment.Counterparty,
		}

		// Шукаємо рахунок у даних контрагента (оригінальний рахунок тощо)
		foundAccount := reader.FindAccountInCounterparty(payment.Counterparty, accountList)
		if foundAccount != "" {
			// Знайшли рахунок - замінюємо контрагента на ПІБ
			fullName, exists := accountToName[foundAccount]
			if exists {
				result.Name = fullName
				result.Account = foundAccount
				result.Counterparty = fullName
				fmt.Printf("Замінено рахунок %s на ПІБ: %s\n", foundAccount, fullName)
			}
		}

		results = append(results, result)
	}

	return results
}
