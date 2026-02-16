// Package model визначає структури даних, які використовуються в застосунку.
package model

// AccrualRecord представляє запис у файлі нарахування з ПІБ та рахунком
type AccrualRecord struct {
	FullName string // ПІБ працівника
	Account  string // Особистий рахунок
}

// PaymentRecord представляє запис платежу з банківського виписку
type PaymentRecord struct {
	Date           string  // Дата платежу
	Sum            float64 // Сума платежу
	Counterparty    string  // ПІБ або рахунок контрагента
	OriginalData   string  // Оригінальні дані контрагента для логування
}

// ResultRecord представляє фінальний результат з замінями
type ResultRecord struct {
	Date        string  // Дата платежу
	Sum         float64 // Сума платежу
	Name        string  // ПІБ (замість рахунку, якщо знайдено)
	Account     string  // Особистий рахунок (якщо знайдено)
	Counterparty string  // ПІБ або рахунок контрагента (оновлений)
}
