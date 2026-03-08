# Code Review — bank-analyzer

> Стек: Go 1.25, Fyne v2.7.3, excelize v2, shopspring/decimal  
> Дата: 2026-03-07

---

## 1. Коректність та надійність

### [internal/ui/categories_screen.go](file:///home/min/git-workspace/go-from-to/internal/ui/categories_screen.go)

**🔴 Race condition — прямий доступ без mutex (рядок 212)**

```go
for _, tx := range state.Transactions {   // ← без RLock!
    if tx.Category == selectedCategory {
        tx.Category = models.UncategorizedCategory
    }
}
```

Цей блок виконується в UI-потоці (колбек `dialog.ShowConfirm`), але
`state.AppendTransactions` може одночасно викликатися з горутини імпорту.
`state.Transactions` — незахищений зріз.

**Виправлення:**
```go
txs := state.GetTransactions()
for _, tx := range txs {
    if tx.Category == selectedCategory {
        tx.Category = models.UncategorizedCategory
    }
}
```
Або винести мутацію в захищений метод `AppState.RecategorizeTo(from, to string)`.

---

**⚠️ `refreshSummary` читає `state.Transactions` без mutex (import_screen.go:92)**

```go
total := len(state.Transactions)  // читається без RLock
```

Під час паралельного імпорту це може дати невірну/stale довжину.  
**Виправлення:** замінити на `len(state.GetTransactions())`.

---

**⚠️ [ParseDecimal](file:///home/min/git-workspace/go-from-to/internal/parsers/utils/utils.go#28-46) — неоднозначність формату `1.234` (utils/utils.go:35-43)**

Рядок `"1.234"` (без коми) трактується як `1.234` (одна ціла двісті тридцять
чотири сотих). Але деякі банки можуть виводити тисячі через крапку (`1.234` =
1234 грн). Логіка не задокументована — у ПриватБанку CSVсумісна, але варто
додати коментар для ясності.

---

**✅ [isSummaryRow](file:///home/min/git-workspace/go-from-to/internal/parsers/xlsx/pb_parser.go#190-202)** — перевіряє лише першу клітинку через `HasPrefix` зі
списком маркерів. Ризик помилкового спрацювання мінімальний.

**✅ Порожній файл / лише заголовки** — обидва парсери повертають осмислену помилку.

**✅ Нульова сума** — CSV відхиляє `amt.IsZero()`; XLSX відхиляє якщо обидва
debit/credit нульові.

**✅ Невалідна дата** — [ParseDate](file:///home/min/git-workspace/go-from-to/internal/parsers/utils/utils.go#12-27) перебирає 6 форматів, повертає `Errorf`.

---

### [internal/ui/app.go](file:///home/min/git-workspace/go-from-to/internal/ui/app.go)

**⚠️ [showProvidersWarning](file:///home/min/git-workspace/go-from-to/internal/ui/app.go#190-213) — мертвий код (рядки 192–212)**

Функція не викликається ніде — замінена [showStartupWarnings](file:///home/min/git-workspace/go-from-to/internal/ui/app.go#157-189). Варто видалити.

---

### [internal/config/config.go](file:///home/min/git-workspace/go-from-to/internal/config/config.go)

**⚠️ Подвійний виклик [ValidateReport311Config](file:///home/min/git-workspace/go-from-to/internal/mappings/report_config.go#59-85) (рядки 45 і 48)**

```go
cfg.Warnings = mappings.ValidateReport311Config(...)  // рядок 45
if warnings := mappings.ValidateReport311Config(...); ...  // рядок 48 — зайвий!
```

Функція викликається двічі. Другий виклик нічого не зберігає і лише дублює
логування. **Виправлення:** видалити блок [if](file:///home/min/git-workspace/go-from-to/internal/ui/app.go#56-62) на рядках 47–52.

---

## 2. Консистентність між файлами

**✅ Wrapped errors через `%w`** — всі парсери та mappings консистентні.

**✅ `log.Printf` у парсерах** — єдиний рівень деталізації, єдиний формат.

**✅ [normalize()](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#125-135) застосовується консистентно** — при завантаженні правил
([buildKeywordMaps](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#44-65)) і при пошуку ([Categorize](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#66-86), [resolveVnesky](file:///home/min/git-workspace/go-from-to/internal/categorizer/provider_resolver.go#53-65)). Ключові
слова нормалізуються при побудові кешу, а не при кожному порівнянні — правильно.

**⚠️ Дублювання [buildIndex](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go#196-203) / [isEmptyRow](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go#204-212)**

- [csv/privatbank.go](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go) — [buildIndex](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go#196-203), [isEmptyRow](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go#204-212)
- [xlsx/pb_parser.go](file:///home/min/git-workspace/go-from-to/internal/parsers/xlsx/pb_parser.go) — [buildXLSXIndex](file:///home/min/git-workspace/go-from-to/internal/parsers/xlsx/pb_parser.go#173-180), [isXLSXEmptyRow](file:///home/min/git-workspace/go-from-to/internal/parsers/xlsx/pb_parser.go#181-189)

Логіка ідентична. Можна перенести до `parsers/utils` — некритично за розміром.

---

## 3. Архітектура та дизайн

**✅ Розподіл відповідальності** — [buildReport](file:///home/min/git-workspace/go-from-to/internal/ui/report_screen.go#496-503) у [report_screen.go](file:///home/min/git-workspace/go-from-to/internal/ui/report_screen.go) лише
делегує до `reports.BuildReport`. Агрегація — в [reports/builder.go](file:///home/min/git-workspace/go-from-to/internal/reports/builder.go). ✅

**✅ [AddTransactionListener](file:///home/min/git-workspace/go-from-to/internal/ui/app.go#34-37) / pubsub** — кожен екран підписується один раз
при `New*Screen`. Екрани створюються один раз у [Run()](file:///home/min/git-workspace/go-from-to/internal/ui/app.go#63-156) — подвійної підписки немає.

**✅ [categories_screen.go](file:///home/min/git-workspace/go-from-to/internal/ui/categories_screen.go) — Uncategorized після видалення** — транзакції
переводяться (рядки 212–216). Але без mutex (⚠️ зазначено в розділі 1).

**⚠️ [BuildReport](file:///home/min/git-workspace/go-from-to/internal/reports/builder.go#10-77) — Percent рахується лише по витратах**

```go
pct, _ := cs.Total.Div(totalExpense)...
```

Категорії типу `credit` (Внески, Контрагенти) теж потрапляють до `ByCategory`
і діляться на `totalExpense`. Відсоток для надходжень буде > 100% якщо їхня
сума більша за витрати. Якщо звіт показує лише debit-категорії — ОК; якщо mixed — ⚠️.

---

## 4. Безпека та стабільність UI

**⚠️ `AllWindows()[0]` — потенційна паніка**

У трьох місцях (`import_screen.go:42`, `categories_screen.go:17`, `report_screen.go:36`):

```go
win := fyne.CurrentApp().Driver().AllWindows()[0]
```

Якщо Fyne ще не створив вікна (тест, race) — паніка. На практиці виклики
відбуваються після `a.NewWindow()`, тому безпечно, але підхід крихкий.

**Виправлення:** передавати `fyne.Window` як параметр у конструктори екранів.

---

**✅ `generateBtn.Disable()` / `Enable()`** — `runInBackground` вимикає обидві
кнопки до завершення та вмикає у `fyne.Do`. Подвійний клік захищений.

---

**⚠️ `runInBackground` — `fyne.Do` з UI-потоку (report_screen.go:167)**

```go
fyne.Do(func() {          // викликається з UI-thread
    progress.Show()
    statusLabel.SetText("Генерація...")
})
go func() { ... }()
```

У Fyne v2 `fyne.Do` з UI-thread виконується синхронно (безпечно), але
це неочевидна залежність від внутрішньої реалізації. Прямий виклик без
`fyne.Do` був би чистішим.

---

**✅ Захист індексів таблиці** — `id.Row-1 >= len(cachedTxs)` та
`id >= len(files)` перевіряються скрізь.

---

## 5. Продуктивність

**✅ [normalize()](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#125-135) — один раз при завантаженні правил** — [buildKeywordMaps](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#44-65)
будується в конструкторі. При категоризації [normalize(desc)](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#125-135) — один раз на транзакцію.

**✅ [filtered()](file:///home/min/git-workspace/go-from-to/internal/ui/transactions_screen.go#174-196) — кешований результат** — `cachedTxs` оновлюється тільки
при зміні фільтрів; рендер таблиці не перераховує фільтр.

**✅ `excelize.OpenFile` закривається через `defer f.Close()`** — в [CanParse](file:///home/min/git-workspace/go-from-to/internal/parsers/pdf/pb_parser.go#17-21)
і [Parse](file:///home/min/git-workspace/go-from-to/internal/parsers/pdf/pb_parser.go#22-26) у XLSX-парсері.

**⚠️ [normalize](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#125-135) — `transform.Chain` створюється при кожному виклику (rules.go:128)**

```go
func normalize(s string) string {
    t := transform.Chain(norm.NFKD, runes.Remove(...))  // новий об'єкт щоразу
    ...
}
```

При категоризації тисяч транзакцій — тисячі тимчасових об'єктів.  
**Виправлення:** зробити трансформер package-level singleton або `sync.Pool`.

---

## 6. YAML конфігурація

**🔴 [categories.yaml](file:///home/min/git-workspace/go-from-to/assets/categories.yaml) — занадто загальне ключове слово `"оплата"` у `Комунальні`**

```yaml
Комунальні:
  type: debit
  keywords:
    - оплата    # ← КРИТИЧНО загальне слово!
```

Слово «оплата» міститься в описі практично будь-якої деbet-транзакції. Це
призведе до масової хибної класифікації більшості транзакцій як «Комунальні».

**Виправлення:** видалити `оплата` із `Комунальні` або замінити на специфічні
підрядки (`оплата за воду`, `ком. послуги`, тощо).

---

**⚠️ [categories.yaml](file:///home/min/git-workspace/go-from-to/assets/categories.yaml) — `Київстар` та `Контрагенти` мають `type: credit`**

`credit` = надходження на рахунок. Якщо ми платимо Київстару — тип має бути
`debit`. Перевірте що відповідає реальному напрямку транзакцій.

---

**✅ `Зарплата: type: debit`** — семантично правильно: це зарплатні витрати
(ПДФО, ЄСВ перераховуються з рахунку). Назва може вводити в оману.

**✅ [column_mappings.yaml](file:///home/min/git-workspace/go-from-to/assets/column_mappings.yaml)** — покриває множинні варіанти назв колонок.
Наявні `debit`/`credit` дають підтримку форматів з окремими стовпцями.

**✅ [report_311.yaml](file:///home/min/git-workspace/go-from-to/assets/report_311.yaml)** — всі категорії присутні у [categories.yaml](file:///home/min/git-workspace/go-from-to/assets/categories.yaml).
[ValidateReport311Config](file:///home/min/git-workspace/go-from-to/internal/mappings/report_config.go#59-85) підтверджує це при кожному старті.

---

## Загальний висновок

| Критерій | Оцінка |
|---|---|
| Коректність | ⚠️ 1 race condition, 1 мертвий код |
| Консистентність | ✅ висока, є мінімальне дублювання |
| Архітектура | ✅ SRP дотримано, pubsub правильний |
| UI-безпека | ⚠️ `AllWindows()[0]` крихко |
| Продуктивність | ⚠️ [normalize](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#125-135) створює трансформер щоразу |
| YAML | 🔴 `"оплата"` у `Комунальні` — хибна класифікація |

### Обов'язкові виправлення перед production

1. **🔴 Видалити `"оплата"` з ключових слів категорії `Комунальні`** (`assets/categories.yaml:10`)
2. **🔴 Захистити `state.Transactions` у `categories_screen.go:212`** — race condition
3. **⚠️ Видалити дублюючий виклик [ValidateReport311Config](file:///home/min/git-workspace/go-from-to/internal/mappings/report_config.go#59-85) у `config.go:47-52`**
4. **⚠️ Замінити `state.Transactions` на `state.GetTransactions()` у `import_screen.go:92`**
5. **⚠️ Видалити мертву функцію [showProvidersWarning](file:///home/min/git-workspace/go-from-to/internal/ui/app.go#190-213) у `app.go:192-212`**

### Некритичні покращення

- Перенести `transformer` у [normalize](file:///home/min/git-workspace/go-from-to/internal/categorizer/rules.go#125-135) до package-level `sync.Pool`
- Передавати `fyne.Window` як параметр у `New*Screen` замість `AllWindows()[0]`
- Перевірити семантику типів для Київстар/Контрагенти у [categories.yaml](file:///home/min/git-workspace/go-from-to/assets/categories.yaml)
- Перевірити розрахунок `Percent` у [BuildReport](file:///home/min/git-workspace/go-from-to/internal/reports/builder.go#10-77) для mixed-категорій
- Об'єднати [buildIndex](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go#196-203)/[isEmptyRow](file:///home/min/git-workspace/go-from-to/internal/parsers/csv/privatbank.go#204-212) у `parsers/utils`
