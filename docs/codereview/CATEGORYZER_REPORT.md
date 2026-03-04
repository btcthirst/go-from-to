# Code Review — `internal/categorizer` + `internal/reports`

> Дата: 2026-03-03
> Ревізор: Claude (Sonnet 4.6)
> Scope: `categorizer/rules.go`, `categorizer/provider_resolver.go`, `reports/builder.go`, `reports/xlsx_report.go`, `reports/xlsx_report_311.go`, `reports/ods_report.go`

---

## Легенда

| Позначка | Значення |
|----------|----------|
| ✅ | Все коректно |
| ⚠️ | Потенційна проблема — варто виправити |
| 🔴 | Критична помилка — обов'язково виправити перед production |

---

## `internal/categorizer/rules.go`

### ✅ Нормалізація консистентна

`normalize()` застосовується і до ключових слів у `buildKeywordMaps()`, і до тексту транзакцій у `Categorize()`. Логіка type-aware через окремі кеші `debitKeywords` / `creditKeywords` — коректна.

---

### ⚠️ `AddRule` зберігає ненормалізоване ключове слово в `rule.Keywords`

**Файл:** `rules.go` → метод `AddRule`

**Проблема:** нормалізований рядок `kw` використовується для кешу, але в `rule.Keywords` записується оригінальний `keyword`. При наступному перезавантаженні конфігу з YAML це ключове слово не буде нормалізоване при `buildKeywordMaps`.

```go
// ❌ Поточний код
rule.Keywords = append(rule.Keywords, keyword)

// ✅ Виправлення
rule.Keywords = append(rule.Keywords, kw) // kw = normalize(keyword)
```

---

### ⚠️ `CategorizeAll` не перекатегоризує `"Uncategorized"`

**Проблема:** функція пропускає транзакції з будь-якою непорожньою категорією, включно з `"Uncategorized"`. Якщо користувач видалив категорію — транзакція «застрягає» і повторна категоризація її не торкнеться.

**Рекомендація:** якщо це навмисна поведінка — задокументувати коментарем. Якщо ні — змінити умову:

```go
// Пропускаємо тільки якщо категорія встановлена вручну або не є fallback
if tx.Category != "" && tx.Category != "Uncategorized" {
    continue
}
```

---

## `internal/categorizer/provider_resolver.go`

### ⚠️ `resolveKommunal` і `resolveVnesky` не нормалізують рядок пошуку

**Проблема:** `RulesCategorizer` нормалізує текст перед порівнянням, але `ProviderResolver` — ні. Якщо в описі транзакції є омогліфи або діакритика, пошук не спрацює.

```go
// ❌ Поточний код
haystack := strings.ToLower(tx.Description + " " + tx.Counterparty)

// ✅ Виправлення
haystack := normalize(tx.Description + " " + tx.Counterparty)
// і для ключів: strings.Contains(haystack, normalize(key))
```

Це стосується обох методів: `resolveKommunal` та `resolveVnesky`.

---

### ⚠️ `communalKeywords` — хардкод поза конфігурацією

**Проблема:** підпостачальники для категорії `"Комунальні"` (`вувкг`, `парки`, `хоек`) захардкоджені у коді. Додати нового постачальника можна тільки через зміну коду та перекомпіляцію. Також ці ключі не нормалізовані.

**Рекомендація:** перенести у `providers.yaml` або окрему секцію `categories.yaml`.

---

### ⚠️ `categoryProvider` дублює назви категорій з `categories.yaml`

**Проблема:** рядки `"Військовий"`, `"ПДФО"`, `"ЄСВ"` і т.д. жорстко вшиті у `map[string]string`. При перейменуванні категорії в YAML — resolver перестане їх знаходити без жодного попередження або помилки компіляції.

**Рекомендація:** додати валідацію при старті або завантажувати маппінг з конфігурації.

---

## `internal/reports/builder.go`

### ✅ Загальна коректність

Чисті функції без side-effects, `shopspring/decimal` скрізь, `Period` обчислюється коректно через `Before`/`After` по всіх транзакціях.

---

### ⚠️ Відсотки `ByCategory` рахуються від суми витрат для всіх категорій

**Проблема:** `ByCategory` містить і credit-категорії (`"Внески"`, `"Контрагенти"`), але їх відсоток ділиться на `totalExpense`. Результат — некоректний `Percent` для категорій надходжень.

```go
// ❌ Поточний код — ділить усі категорії на totalExpense
pct, _ := cs.Total.Div(totalExpense).Mul(decimal.NewFromInt(100)).Float64()
```

**Рекомендація:** рахувати відсоток окремо для debit-категорій (від `totalExpense`) і credit-категорій (від `totalIncome`), або взагалі виключити credit-категорії з `ByCategory`.

---

## `internal/reports/xlsx_report.go`

### 🔴 `NumFmt: 177` — нестандартний числовий формат

**Файл:** `xlsx_report.go` → метод `writeTransactions`

**Проблема:** числові формати 164+ є кастомними слотами, залежними від реалізації. `177` може відображатись некоректно або взагалі не розпізнаватись у LibreOffice Calc. У `writeDTOs` ця проблема вже вирішена — використовується `NumFmt: 4`.

```go
// ❌ Поточний код
incomeStyle, _ := f.NewStyle(&excelize.Style{
    NumFmt: 177,
})

// ✅ Виправлення
customFmt := `#,##0.00`
incomeStyle, _ := f.NewStyle(&excelize.Style{
    CustomNumFmt: &customFmt,
})
```

---

### ⚠️ `writeTransactions` застосовує `SetRowStyle` — числові колонки отримують текстовий стиль

**Проблема:** `SetRowStyle` накидає один стиль на весь рядок. Колонка «Сума» буде мати той самий стиль що й текстові поля, а `NumFmt` у рядковому стилі може конфліктувати з числовими значеннями. У `writeDTOs` ця проблема вирішена поколонковим підходом.

**Рекомендація:** встановлювати стиль поколонково, як у `writeDTOs`.

---

### ⚠️ `Generate` не видаляє дефолтний `Sheet1`

**Проблема:** `GenerateFromDTO` видаляє `Sheet1` при створенні нового файлу, але `Generate` — ні. У результаті у повному звіті буде зайвий порожній лист.

```go
// Додати на початку Generate, аналогічно до writeDTOs:
if idx, err := f.GetSheetIndex("Sheet1"); err == nil && idx >= 0 {
    f.DeleteSheet("Sheet1")
}
```

---

## `internal/reports/xlsx_report_311.go`

### ⚠️ `AutoFilter` захоплює рядок «РАЗОМ»

**Проблема:** `apply311Settings` передає `totalRow` (рядок підсумків) як межу AutoFilter. Це означає що рядок «РАЗОМ» потрапляє у фільтрований діапазон і може сховатись при застосуванні фільтру.

```go
// ❌ Поточний код
f.AutoFilter(sheet, fmt.Sprintf("A2:%s%d", lastCol, totalRow), nil)

// ✅ Виправлення — передавати lastData замість totalRow
f.AutoFilter(sheet, fmt.Sprintf("A2:%s%d", lastCol, lastData), nil)
```

Для цього треба змінити сигнатуру `apply311Settings`:

```go
func apply311Settings(f *excelize.File, sheet string, lastData, totCol int)
```

---

### ⚠️ Анонімна функція у `switch` — нечитабельний патерн

**Проблема:** перевірка наявності ключа в `subIdx` через самовикликувану функцію — нестандартний Go-патерн, що ускладнює читання коду.

```go
// ❌ Поточний код
case func() bool { _, ok := subIdx[tx.Category]; return ok }():

// ✅ Виправлення — підняти перевірку до switch
subColIdx, isSubCol := subIdx[tx.Category]
switch {
case mainSet[tx.Category]:
    // ...
case isSubCol:
    idx := subColIdx
    // ...
}
```

---

## `internal/reports/ods_report.go`

### ⚠️ `buildDTORows` не включає поле `Provider`

**Проблема:** `buildTransactionRows` виводить `tx.Provider`, але `buildDTORows` — ні, хоча `TransactionDTO` містить поле `Provider`. Непослідовність між XLSX і ODS DTO-експортом.

---

### ⚠️ Нейтральні рядки підсумків стилізуються як `"ceIncome"` (зелений фон)

**Проблема:** рядки з `buildSummaryRows` мають `debit: false` → стиль `ceIncome`. Рядок «Витрати» у підсумку відображається зеленим — візуально оманливо.

**Рекомендація:** додати нейтральний стиль `ceDefault` у `template.ods` та повертати його для рядків підсумків.

---

## Зведена таблиця проблем

| # | Проблема | Файл | Критичність |
|---|----------|------|-------------|
| 1 | `NumFmt: 177` — зламає відображення у LibreOffice | `xlsx_report.go` | 🔴 |
| 2 | `resolveKommunal`/`resolveVnesky` без нормалізації — омогліфи не знайдуться | `provider_resolver.go` | ⚠️ |
| 3 | AutoFilter захоплює рядок «РАЗОМ» у 311-звіті | `xlsx_report_311.go` | ⚠️ |
| 4 | `AddRule` зберігає ненормалізоване ключове слово | `rules.go` | ⚠️ |
| 5 | Відсотки credit-категорій рахуються від суми витрат | `builder.go` | ⚠️ |
| 6 | `Sheet1` не видаляється у повному `Generate` | `xlsx_report.go` | ⚠️ |
| 7 | `communalKeywords` — хардкод поза конфігурацією | `provider_resolver.go` | ⚠️ |
| 8 | `categoryProvider` дублює назви категорій з YAML | `provider_resolver.go` | ⚠️ |
| 9 | `Provider` відсутній у DTO ODS-експорті | `ods_report.go` | ⚠️ |
| 10 | `SetRowStyle` не розрізняє числові та текстові колонки | `xlsx_report.go` | ⚠️ |
| 11 | Анонімна функція у `switch` — нечитабельний патерн | `xlsx_report_311.go` | ⚠️ |
| 12 | Нейтральні рядки підсумків в ODS стилізовані як credit | `ods_report.go` | ⚠️ |

---

## Висновок

**Обов'язково перед production (блокуючі):**

1. **`NumFmt: 177`** → замінити на `CustomNumFmt: &customFmt` з `"#,##0.00"`. Поточний код генерує XLSX-файли що некоректно відкриваються в LibreOffice.
2. **Нормалізація в `ProviderResolver`** → без цього пошук постачальників не спрацює для транзакцій з омогліфами або мішаним регістром — а це основна функціональність звіту 311.
3. **AutoFilter у `xlsx_report_311.go`** → рядок «РАЗОМ» зникатиме при будь-якій фільтрації у Excel.

**Важливо, але не блокує запуск:**

Проблеми 4–12 впливають на коректність даних, підтримуваність та консистентність UI, але не призведуть до краш-помилок при першому запуску. Рекомендується виправити до передачі у регулярне використання.