# TASKS.md — Bank Statement Analyzer

## Легенда
- `[ ]` — не розпочато
- `[~]` — в процесі
- `[x]` — завершено

Пріоритет: 🔴 критично → 🟡 важливо → 🟢 бажано

---

## 🏗️ Milestone 1 — Фундамент проєкту

### Ініціалізація
- [x] 🔴 Створити `go.mod` з назвою модуля `bank-analyzer`
- [x] 🔴 Додати залежності: `fyne/v2`, `excelize/v2`, `shopspring/decimal`, `golang.org/x/text`, `gopkg.in/yaml.v3`, `pdfcpu`
- [x] 🔴 Створити структуру директорій згідно з архітектурним документом
- [x] 🔴 Налаштувати `cmd/main.go` — точка входу, запуск Fyne app

### Моделі даних
- [x] 🔴 `internal/models/transaction.go` — структура `Transaction`, тип `TransactionType` (Debit/Credit), метод `NetAmount()`
- [x] 🔴 `internal/models/report.go` — структури `Report`, `CategorySummary`, `MonthSummary`, `DateRange`

---

## 🔌 Milestone 2 — Парсери

### Інфраструктура парсерів
- [x] 🔴 `internal/parsers/parser.go` — інтерфейс `Parser` з методами `Name()`, `CanParse()`, `Parse()`
- [x] 🔴 `internal/parsers/registry.go` — `Registry`, методи `Register()`, `Detect()`, `ParseFile()`

### CSV парсери
- [ ] 🔴 `internal/parsers/csv/privatbank.go`
  - [ ] Детекція по заголовках (Windows-1251, роздільник `;`)
  - [ ] Парсинг дати формату `dd.mm.yyyy`
  - [ ] Визначення Debit/Credit по знаку суми
  - [ ] Маппінг полів: дата, сума, валюта, опис, контрагент
- [ ] 🔴 `internal/parsers/csv/monobank.go`
  - [ ] Детекція по заголовках (UTF-8, роздільник `,`)
  - [ ] Парсинг дати формату `dd.mm.yyyy HH:MM:SS`
  - [ ] Маппінг специфічних полів Monobank (MCC код тощо)
- [ ] 🟡 `internal/parsers/csv/oschadbank.go`
  - [ ] Дослідити формат виписки Ощадбанку
  - [ ] Реалізувати парсер
- [ ] 🟡 `internal/parsers/csv/pumb.go`
  - [ ] Дослідити формат виписки ПУМБ
  - [ ] Реалізувати парсер

### XLSX парсер (вхідний)
- [ ] 🟡 `internal/parsers/xlsx/xlsx_parser.go`
  - [ ] Детекція формату по першому листу та заголовках
  - [ ] Парсинг через `excelize.OpenFile()`
  - [ ] Підтримка різних розкладок колонок

### PDF парсер
- [ ] 🟢 `internal/parsers/pdf/pdf_parser.go` — базова структура
- [ ] 🟢 `internal/parsers/pdf/table_extractor.go` — витягування таблиць через `pdfcpu`
- [ ] 🟢 Тестування на реальних PDF-виписках ПриватБанку / Ощадбанку

### Тести парсерів
- [ ] 🔴 Додати тестові файли виписок у `testdata/` (знеособлені)
- [ ] 🔴 Unit-тести для `PrivatBankParser`
- [ ] 🔴 Unit-тести для `MonobankParser`
- [ ] 🟡 Unit-тести для `Registry.Detect()`
- [ ] 🟡 Тест на коректність `decimal` сум (без float-похибок)

---

## 🏷️ Milestone 3 — Категоризатор

- [ ] 🔴 `assets/categories.yaml` — базові правила (Продукти, Транспорт, Комуналка, ЗП, Медицина, Розваги, Кешбек)
- [ ] 🔴 `internal/categorizer/categorizer.go` — інтерфейс `Categorizer`
- [ ] 🔴 `internal/categorizer/rules.go`
  - [ ] Завантаження YAML конфігу
  - [ ] Побудова кешу `keywordMap`
  - [ ] Метод `Categorize(tx)` — пошук по опису та контрагенту
  - [ ] Метод `CategorizeAll(txs)`
  - [ ] Метод `AddRule(category, keyword)` — додавання правил з UI
  - [ ] Метод `Save(path)` — збереження оновленого YAML
- [ ] 🟡 Врахування `type` (debit/credit) та `min_amount` у правилах
- [ ] 🟢 `internal/categorizer/ml.go` — заглушка для майбутньої ML-категоризації

---

## ⚙️ Milestone 4 — Конфігурація

- [ ] 🔴 `internal/config/config.go`
  - [ ] Структура `Config` (шлях до categories.yaml, останній використаний каталог, тема)
  - [ ] `Load()` — зчитування через `fyne.App.Preferences()`
  - [ ] `Save()` — збереження налаштувань
- [ ] 🔴 `internal/config/categories.go`
  - [ ] `CategoryNames()` — список назв для UI-селекторів
  - [ ] `AddCategory(name)` / `RemoveCategory(name)`

---

## 📊 Milestone 5 — Генерація звітів

### XLSX
- [ ] 🔴 `internal/reports/builder.go` — інтерфейс `ReportBuilder` з методом `Generate(report, outputPath) error`
- [ ] 🔴 `internal/reports/xlsx_report.go`
  - [ ] `writeTransactions()` — лист з усіма транзакціями, стилі, автофільтр, заморожений рядок
  - [ ] `writeSummary()` — лист з підсумковими показниками
  - [ ] `writePivot()` — лист "За категоріями" з сумами та відсотками
  - [ ] `writeMonthly()` — лист "По місяцях"
  - [ ] `writeMonthlyChart()` — стовпчиковий графік руху коштів
  - [ ] Підтримка шаблону: `excelize.OpenFile()` якщо `TemplatePath` вказано
  - [ ] Коректне підсвічування витрат (червоний) та надходжень (зелений)
  - [ ] Числовий формат клітинок для сум `#,##0.00`
- [ ] 🟡 `internal/reports/ods_report.go` — аналогічна структура для ODS формату

### Побудова звіту
- [ ] 🔴 `buildReport(state)` — агрегація транзакцій у `models.Report`
  - [ ] Підрахунок `TotalIncome`, `TotalExpense`, `NetBalance`
  - [ ] Групування `ByCategory`
  - [ ] Групування `ByMonth` (ключ `"2024-01"`)
  - [ ] Визначення `Period` (мін/макс дата)

---

## 🖥️ Milestone 6 — UI

### Головне вікно
- [ ] 🔴 `internal/ui/app.go`
  - [ ] Ініціалізація `AppState`
  - [ ] Бокова навігація (`widget.List`)
  - [ ] `container.NewHSplit` з навігацією та контентом (offset 0.2)
  - [ ] Перемикання екранів

### Екран імпорту
- [ ] 🔴 `internal/ui/import_screen.go`
  - [ ] Кнопка "Додати файл" з `dialog.NewFileOpen`, фільтр розширень
  - [ ] Список завантажених файлів з іконками та назвою банку
  - [ ] Запуск парсингу в горутині, без блокування UI
  - [ ] Показ помилок через `dialog.ShowError`
  - [ ] Оновлення лічильника транзакцій після імпорту
- [ ] 🟡 `internal/ui/widgets/file_dropper.go` — Drag & Drop зона

### Екран транзакцій
- [ ] 🔴 `internal/ui/transactions_screen.go`
  - [ ] `widget.Table` з 7 колонками
  - [ ] Рядок фільтрів: пошук по тексту, вибір категорії, вибір типу
  - [ ] Функція `filtered()` — фільтрація транзакцій по активних фільтрах
  - [ ] Налаштування ширини колонок
  - [ ] Статус-рядок з кількістю знайдених транзакцій
  - [ ] `OnSelected` — відкриття діалогу редагування категорії транзакції
- [ ] 🟡 `showCategoryEditor()` — діалог зміни категорії + опція "додати правило"

### Екран категорій
- [ ] 🟡 `internal/ui/categories_screen.go`
  - [ ] Список категорій з кількістю ключових слів
  - [ ] Редагування ключових слів категорії
  - [ ] Додавання нової категорії
  - [ ] Збереження змін у YAML

### Екран звіту
- [ ] 🔴 `internal/ui/report_screen.go`
  - [ ] Вибір формату (XLSX / ODS)
  - [ ] Кнопка вибору шаблону
  - [ ] Поля фільтру дат "від" / "до"
  - [ ] Чекбокси секцій: Підсумок, За категоріями, По місяцях, Графік
  - [ ] `dialog.NewFileSave` для вибору шляху збереження
  - [ ] Генерація в горутині з `widget.ProgressBar`
  - [ ] Повідомлення про успіх / помилку

---

## 🔧 Milestone 7 — Полірування

- [ ] 🟡 Обробка дублікатів транзакцій при повторному імпорті одного файлу (по хешу дата+сума+опис)
- [ ] 🟡 Підтримка мультивалютності (відображення оригінальної валюти)
- [ ] 🟡 Кнопка "Очистити всі транзакції"
- [ ] 🟡 Збереження/відновлення стану між сесіями (список файлів, категорії транзакцій)
- [ ] 🟢 Темна тема Fyne
- [ ] 🟢 Пряма інтеграція з Monobank API (`api.monobank.ua`)
- [ ] 🟢 Підтримка курсів НБУ для мультивалютних звітів
- [ ] 🟢 Іконка застосунку та `fyne package` для дистрибуції

---

## 📦 Порядок реалізації (рекомендований)

```
M1 (фундамент) → M2 CSV парсери → M3 категоризатор → M5 XLSX звіт
     → M4 конфіг → M6 UI → M2 PDF парсер → M7 полірування
```

Перший робочий прототип можливий після **M1 + M2 (CSV) + M3 + M5** —
тобто без UI, але з можливістю запустити як CLI і отримати XLSX.