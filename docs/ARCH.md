# Bank Statement Analyzer — Архітектура (Go + Fyne)

## Огляд проєкту

Десктоп-застосунок для парсингу банківських виписок різних форматів,
категоризації транзакцій та експорту звітів у XLSX/ODS.

**Стек:** Go 1.22+, Fyne v2, excelize (XLSX), pdfcpu або pdfextract (PDF)

---

## Структура проєкту

```
bank-analyzer/
├── cmd/
│   └── main.go                  # Точка входу, ініціалізація Fyne App
├── internal/
│   ├── models/
│   │   ├── transaction.go       # Структура транзакції
│   │   └── report.go            # Структура звіту
│   ├── parsers/
│   │   ├── parser.go            # Інтерфейс Parser
│   │   ├── registry.go          # Авто-детектор формату
│   │   ├── csv/
│   │   │   ├── privatbank.go    # ПриватБанк CSV
│   │   │   ├── monobank.go      # Monobank CSV
│   │   │   ├── oschadbank.go    # Ощадбанк CSV
│   │   │   └── pumb.go          # ПУМБ CSV
│   │   ├── pdf/
│   │   │   ├── pdf_parser.go    # Базовий PDF парсер
│   │   │   └── table_extractor.go
│   │   └── xlsx/
│   │       └── xlsx_parser.go   # Читання вхідних XLSX
│   ├── categorizer/
│   │   ├── categorizer.go       # Інтерфейс
│   │   ├── rules.go             # Правила на основі ключових слів
│   │   └── ml.go               # (майбутнє) ML-категоризація
│   ├── reports/
│   │   ├── builder.go           # Інтерфейс ReportBuilder
│   │   ├── xlsx_report.go       # Генерація XLSX через excelize
│   │   └── ods_report.go        # Генерація ODS
│   ├── config/
│   │   ├── config.go            # Завантаження/збереження налаштувань
│   │   └── categories.go        # Управління категоріями
│   └── ui/
│       ├── app.go               # Головне вікно, навігація
│       ├── import_screen.go     # Екран імпорту файлів
│       ├── transactions_screen.go # Таблиця транзакцій
│       ├── categories_screen.go  # Редактор категорій
│       ├── report_screen.go     # Налаштування та генерація звіту
│       └── widgets/
│           ├── file_dropper.go  # Drag & drop зона
│           ├── tx_table.go      # Таблиця транзакцій з фільтром
│           └── chart_widget.go  # Кастомний віджет графіку
├── assets/
│   └── categories.yaml          # Правила категоризації за замовчуванням
├── go.mod
└── go.sum
```

---

## Моделі даних

### `internal/models/transaction.go`

```go
package models

import (
    "time"
    "github.com/shopspring/decimal"
)

type TransactionType int

const (
    Debit  TransactionType = iota // витрата
    Credit                        // надходження
)

type Transaction struct {
    ID          string
    Date        time.Time
    Amount      decimal.Decimal // завжди позитивне
    Type        TransactionType // Debit або Credit
    Currency    string
    Description string
    Counterparty string          // контрагент якщо є
    IBAN        string
    Category    string
    Balance     decimal.Decimal  // залишок після операції
    BankSource  string           // "privatbank", "monobank", etc.
    Raw         map[string]string // оригінальні поля
}

// NetAmount повертає зі знаком: від'ємне для Debit
func (t *Transaction) NetAmount() decimal.Decimal {
    if t.Type == Debit {
        return t.Amount.Neg()
    }
    return t.Amount
}
```

### `internal/models/report.go`

```go
package models

import "github.com/shopspring/decimal"

type Report struct {
    Transactions  []*Transaction
    Period        DateRange
    TotalIncome   decimal.Decimal
    TotalExpense  decimal.Decimal
    NetBalance    decimal.Decimal
    ByCategory    map[string]CategorySummary
    ByMonth       map[string]MonthSummary  // ключ: "2024-01"
}

type CategorySummary struct {
    Category string
    Count    int
    Total    decimal.Decimal
    Percent  float64
}

type MonthSummary struct {
    Month   string
    Income  decimal.Decimal
    Expense decimal.Decimal
}

type DateRange struct {
    From time.Time
    To   time.Time
}
```

---

## Парсери

### Інтерфейс (`internal/parsers/parser.go`)

```go
package parsers

import "bank-analyzer/internal/models"

// Parser — інтерфейс для всіх парсерів банківських виписок
type Parser interface {
    // Name повертає назву банку/формату
    Name() string
    // CanParse визначає чи може парсер обробити цей файл
    // (перевіряє розширення та внутрішній вміст)
    CanParse(filepath string) (bool, error)
    // Parse зчитує файл і повертає транзакції
    Parse(filepath string) ([]*models.Transaction, error)
}
```

### Реєстр та авто-детекція (`internal/parsers/registry.go`)

```go
package parsers

import (
    "fmt"
    "bank-analyzer/internal/parsers/csv"
    "bank-analyzer/internal/parsers/pdf"
    xlsxp "bank-analyzer/internal/parsers/xlsx"
)

type Registry struct {
    parsers []Parser
}

func NewRegistry() *Registry {
    r := &Registry{}
    // Реєструємо парсери — порядок важливий (специфічні перед загальними)
    r.Register(csv.NewPrivatBankParser())
    r.Register(csv.NewMonobankParser())
    r.Register(csv.NewOschadbankParser())
    r.Register(csv.NewPUMBParser())
    r.Register(xlsxp.NewParser())
    r.Register(pdf.NewParser())
    return r
}

func (r *Registry) Register(p Parser) {
    r.parsers = append(r.parsers, p)
}

// Detect знаходить підходящий парсер для файлу
func (r *Registry) Detect(filepath string) (Parser, error) {
    for _, p := range r.parsers {
        ok, err := p.CanParse(filepath)
        if err != nil {
            continue
        }
        if ok {
            return p, nil
        }
    }
    return nil, fmt.Errorf("невідомий формат файлу: %s", filepath)
}
```

### Приклад: ПриватБанк CSV (`internal/parsers/csv/privatbank.go`)

```go
package csv

import (
    "encoding/csv"
    "os"
    "strings"
    "time"
    "golang.org/x/text/encoding/charmap"
    "bank-analyzer/internal/models"
    "github.com/shopspring/decimal"
)

// Приклад рядка ПриватБанк:
// "01.03.2024";"Оплата";"Магазин АТБ";"UA123...";"UAH";"-245.50";"10000.00"

type PrivatBankParser struct{}

func NewPrivatBankParser() *PrivatBankParser { return &PrivatBankParser{} }

func (p *PrivatBankParser) Name() string { return "ПриватБанк" }

func (p *PrivatBankParser) CanParse(filepath string) (bool, error) {
    if !strings.HasSuffix(strings.ToLower(filepath), ".csv") {
        return false, nil
    }
    // Читаємо перший рядок для перевірки заголовків
    f, err := os.Open(filepath)
    if err != nil {
        return false, err
    }
    defer f.Close()

    // ПриватБанк використовує Windows-1251
    reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
    reader.Comma = ';'
    header, err := reader.Read()
    if err != nil {
        return false, nil
    }
    // Перевіряємо специфічні заголовки ПриватБанку
    return containsAll(header, []string{"Дата", "Призначення", "Сума"}), nil
}

func (p *PrivatBankParser) Parse(filepath string) ([]*models.Transaction, error) {
    f, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(f))
    reader.Comma = ';'
    reader.LazyQuotes = true

    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    // Визначаємо індекси колонок по заголовку
    header := records[0]
    idx := buildIndex(header)

    var txs []*models.Transaction
    for _, row := range records[1:] {
        if len(row) < 5 {
            continue
        }
        date, err := time.Parse("02.01.2006", row[idx["Дата"]])
        if err != nil {
            continue
        }
        amtStr := strings.ReplaceAll(row[idx["Сума"]], " ", "")
        amt, err := decimal.NewFromString(strings.ReplaceAll(amtStr, ",", "."))
        if err != nil {
            continue
        }
        txType := models.Credit
        if amt.IsNegative() {
            txType = models.Debit
            amt = amt.Abs()
        }
        tx := &models.Transaction{
            Date:         date,
            Amount:       amt,
            Type:         txType,
            Currency:     row[idx["Валюта"]],
            Description:  row[idx["Призначення платежу"]],
            Counterparty: row[idx["Контрагент"]],
            BankSource:   "privatbank",
        }
        txs = append(txs, tx)
    }
    return txs, nil
}
```

---

## Категоризатор

### `internal/categorizer/rules.go`

```go
package categorizer

import (
    "strings"
    "gopkg.in/yaml.v3"
    "os"
    "bank-analyzer/internal/models"
)

type Rule struct {
    Keywords []string `yaml:"keywords"`
    MinAmt   float64  `yaml:"min_amount,omitempty"`
    MaxAmt   float64  `yaml:"max_amount,omitempty"`
    Type     string   `yaml:"type,omitempty"` // "debit", "credit", або ""
}

type CategoryConfig map[string]Rule

type RulesCategorizer struct {
    categories CategoryConfig
    // Кеш для швидкого пошуку
    keywordMap map[string]string // keyword -> category
}

func NewRulesCategorizer(configPath string) (*RulesCategorizer, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    var cfg CategoryConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    c := &RulesCategorizer{categories: cfg, keywordMap: make(map[string]string)}
    // Будуємо кеш
    for cat, rule := range cfg {
        for _, kw := range rule.Keywords {
            c.keywordMap[strings.ToLower(kw)] = cat
        }
    }
    return c, nil
}

func (c *RulesCategorizer) Categorize(tx *models.Transaction) string {
    desc := strings.ToLower(tx.Description + " " + tx.Counterparty)
    for kw, cat := range c.keywordMap {
        if strings.Contains(desc, kw) {
            return cat
        }
    }
    return "Інше"
}

func (c *RulesCategorizer) CategorizeAll(txs []*models.Transaction) {
    for _, tx := range txs {
        if tx.Category == "" {
            tx.Category = c.Categorize(tx)
        }
    }
}

// AddRule дозволяє користувачу додавати правило через UI
func (c *RulesCategorizer) AddRule(category, keyword string) {
    kw := strings.ToLower(keyword)
    c.keywordMap[kw] = category
    rule := c.categories[category]
    rule.Keywords = append(rule.Keywords, kw)
    c.categories[category] = rule
}
```

### `assets/categories.yaml` (вбудовані правила)

```yaml
Продукти харчування:
  type: debit
  keywords:
    - сільпо, атб, novus, fora, metro, ашан, ultramarket
    - продукти, супермаркет, grocery

Транспорт:
  type: debit
  keywords:
    - uber, bolt, uklon, таксі
    - укрзалізниця, поїзд, квиток
    - парковка, автостанція
    - пальне, wog, okko, shell, socar

Комунальні послуги:
  type: debit
  keywords:
    - газ, водоканал, тепло, електро
    - квартплата, жкг, осбб
    - київстар, vodafone, lifecell

Зарплата:
  type: credit
  min_amount: 1000
  keywords:
    - зарплата, зп, salary, виплата
    - нарахування заробітної

Медицина:
  type: debit
  keywords:
    - аптека, pharmacy, лікарня, клініка
    - добробут, medicare, eurolab

Розваги:
  type: debit
  keywords:
    - кіно, театр, concert, netflix, spotify
    - steam, playstation, xbox

Кешбек/Бонуси:
  type: credit
  keywords:
    - cashback, кешбек, бонус, повернення
```

---

## Генерація XLSX-звіту

### `internal/reports/xlsx_report.go`

```go
package reports

import (
    "fmt"
    "github.com/xuri/excelize/v2"
    "bank-analyzer/internal/models"
)

// Стилі кольорів
const (
    colorHeader   = "4472C4"
    colorIncome   = "E2EFDA"
    colorExpense  = "FCE4D6"
    colorSubtotal = "D9E1F2"
)

type XLSXReporter struct {
    TemplatePath string // якщо порожній — створює новий файл
}

func (r *XLSXReporter) Generate(report *models.Report, outputPath string) error {
    var f *excelize.File
    if r.TemplatePath != "" {
        var err error
        f, err = excelize.OpenFile(r.TemplatePath)
        if err != nil {
            return fmt.Errorf("не вдалося відкрити шаблон: %w", err)
        }
    } else {
        f = excelize.NewFile()
    }
    defer f.Close()

    if err := r.writeTransactions(f, report); err != nil {
        return err
    }
    if err := r.writeSummary(f, report); err != nil {
        return err
    }
    if err := r.writePivot(f, report); err != nil {
        return err
    }
    if err := r.writeMonthlyChart(f, report); err != nil {
        return err
    }

    return f.SaveAs(outputPath)
}

func (r *XLSXReporter) writeTransactions(f *excelize.File, report *models.Report) error {
    sheet := "Транзакції"
    f.NewSheet(sheet)

    // Заголовки
    headers := []string{"Дата", "Тип", "Сума", "Валюта", "Категорія", "Опис", "Контрагент", "Баланс"}
    for i, h := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue(sheet, cell, h)
    }

    // Стиль заголовку
    headerStyle, _ := f.NewStyle(&excelize.Style{
        Fill:      excelize.Fill{Type: "pattern", Color: []string{colorHeader}, Pattern: 1},
        Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
        Alignment: &excelize.Alignment{Horizontal: "center"},
    })
    f.SetRowStyle(sheet, 1, 1, headerStyle)

    incomeStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{colorIncome}, Pattern: 1},
        NumFmt: 177, // #,##0.00
    })
    expenseStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{colorExpense}, Pattern: 1},
        NumFmt: 177,
    })

    for i, tx := range report.Transactions {
        row := i + 2
        typeStr := "Надходження"
        style := incomeStyle
        if tx.Type == models.Debit {
            typeStr = "Витрата"
            style = expenseStyle
        }

        data := []interface{}{
            tx.Date.Format("02.01.2006"),
            typeStr,
            tx.Amount.InexactFloat64(),
            tx.Currency,
            tx.Category,
            tx.Description,
            tx.Counterparty,
            tx.Balance.InexactFloat64(),
        }
        for col, val := range data {
            cell, _ := excelize.CoordinatesToCellName(col+1, row)
            f.SetCellValue(sheet, cell, val)
        }
        f.SetRowStyle(sheet, row, row, style)
    }

    // Автофільтр
    f.AutoFilter(sheet, "A1", fmt.Sprintf("H%d", len(report.Transactions)+1), "")
    // Заморозити перший рядок
    f.SetPanes(sheet, &excelize.Panes{
        Freeze:      true,
        Split:       false,
        YSplit:      1,
        TopLeftCell: "A2",
        ActivePane:  "bottomLeft",
    })

    return nil
}

func (r *XLSXReporter) writeSummary(f *excelize.File, report *models.Report) error {
    sheet := "Підсумок"
    f.NewSheet(sheet)

    data := [][]interface{}{
        {"Показник", "Значення"},
        {"Надходження", report.TotalIncome.InexactFloat64()},
        {"Витрати", report.TotalExpense.InexactFloat64()},
        {"Баланс", report.NetBalance.InexactFloat64()},
        {"Транзакцій", len(report.Transactions)},
        {"Період з", report.Period.From.Format("02.01.2006")},
        {"Період по", report.Period.To.Format("02.01.2006")},
    }
    for i, row := range data {
        for j, val := range row {
            cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
            f.SetCellValue(sheet, cell, val)
        }
    }
    return nil
}

func (r *XLSXReporter) writePivot(f *excelize.File, report *models.Report) error {
    sheet := "За категоріями"
    f.NewSheet(sheet)
    f.SetCellValue(sheet, "A1", "Категорія")
    f.SetCellValue(sheet, "B1", "Сума")
    f.SetCellValue(sheet, "C1", "% від витрат")
    f.SetCellValue(sheet, "D1", "Кількість")

    row := 2
    for _, summary := range report.ByCategory {
        f.SetCellValue(sheet, fmt.Sprintf("A%d", row), summary.Category)
        f.SetCellValue(sheet, fmt.Sprintf("B%d", row), summary.Total.InexactFloat64())
        f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", summary.Percent))
        f.SetCellValue(sheet, fmt.Sprintf("D%d", row), summary.Count)
        row++
    }
    return nil
}

func (r *XLSXReporter) writeMonthlyChart(f *excelize.File, report *models.Report) error {
    sheet := "Графік"
    f.NewSheet(sheet)
    // Дані для графіку
    f.SetCellValue(sheet, "A1", "Місяць")
    f.SetCellValue(sheet, "B1", "Надходження")
    f.SetCellValue(sheet, "C1", "Витрати")

    row := 2
    for month, ms := range report.ByMonth {
        f.SetCellValue(sheet, fmt.Sprintf("A%d", row), month)
        f.SetCellValue(sheet, fmt.Sprintf("B%d", row), ms.Income.InexactFloat64())
        f.SetCellValue(sheet, fmt.Sprintf("C%d", row), ms.Expense.InexactFloat64())
        row++
    }

    // Стовпчиковий графік
    chart := &excelize.Chart{
        Type: excelize.Bar,
        Series: []excelize.ChartSeries{
            {Name: sheet + "!$B$1", Categories: fmt.Sprintf("%s!$A$2:$A$%d", sheet, row-1),
             Values: fmt.Sprintf("%s!$B$2:$B$%d", sheet, row-1)},
            {Name: sheet + "!$C$1", Categories: fmt.Sprintf("%s!$A$2:$A$%d", sheet, row-1),
             Values: fmt.Sprintf("%s!$C$2:$C$%d", sheet, row-1)},
        },
        Title: []excelize.RichTextRun{{Text: "Рух коштів по місяцях"}},
    }
    return f.AddChart(sheet, "E2", chart)
}
```

---

## UI (Fyne)

### Головне вікно (`internal/ui/app.go`)

```go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
    "bank-analyzer/internal/parsers"
    "bank-analyzer/internal/categorizer"
    "bank-analyzer/internal/config"
)

type AppState struct {
    Transactions []*models.Transaction
    Registry     *parsers.Registry
    Categorizer  *categorizer.RulesCategorizer
    Config       *config.Config
}

func Run() {
    a := app.NewWithID("ua.bankanalyzer")
    a.Settings().SetTheme(theme.LightTheme())

    w := a.NewWindow("Аналізатор банківських виписок")
    w.Resize(fyne.NewSize(1200, 800))

    state := &AppState{
        Registry:    parsers.NewRegistry(),
        Categorizer: mustLoadCategorizer(),
        Config:      config.Load(),
    }

    // Навігація
    nav := widget.NewList(
        func() int { return 4 },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(id widget.ListItemID, item fyne.CanvasObject) {
            labels := []string{"📂 Імпорт", "📋 Транзакції", "🏷️ Категорії", "📊 Звіт"}
            item.(*widget.Label).SetText(labels[id])
        },
    )

    content := container.NewStack()
    screens := []fyne.CanvasObject{
        NewImportScreen(state, content),
        NewTransactionsScreen(state),
        NewCategoriesScreen(state),
        NewReportScreen(state),
    }

    nav.OnSelected = func(id widget.ListItemID) {
        content.Objects = []fyne.CanvasObject{screens[id]}
        content.Refresh()
    }
    nav.Select(0)

    split := container.NewHSplit(nav, content)
    split.SetOffset(0.2)

    w.SetContent(split)
    w.ShowAndRun()
}
```

### Екран імпорту (`internal/ui/import_screen.go`)

```go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/storage"
    "fyne.io/fyne/v2/widget"
    "bank-analyzer/internal/ui/widgets"
)

func NewImportScreen(state *AppState, content *fyne.Container) fyne.CanvasObject {
    // Список завантажених файлів
    fileList := widget.NewList(
        func() int { return len(state.loadedFiles) },
        func() fyne.CanvasObject {
            return container.NewHBox(
                widget.NewIcon(theme.DocumentIcon()),
                widget.NewLabel(""),
                widget.NewLabel(""),  // статус/банк
            )
        },
        func(id widget.ListItemID, item fyne.CanvasObject) {
            // заповнити рядок списку
        },
    )

    // Кнопка вибору файлу
    addBtn := widget.NewButtonWithIcon("Додати файл", theme.FolderOpenIcon(), func() {
        d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
            if reader == nil {
                return
            }
            go func() {
                txs, parseErr := state.Registry.ParseFile(reader.URI().Path())
                if parseErr != nil {
                    dialog.ShowError(parseErr, fyne.CurrentApp().Driver().AllWindows()[0])
                    return
                }
                state.Categorizer.CategorizeAll(txs)
                state.Transactions = append(state.Transactions, txs...)
                fileList.Refresh()
            }()
        }, fyne.CurrentApp().Driver().AllWindows()[0])
        d.SetFilter(storage.NewExtensionFileFilter([]string{".csv", ".xlsx", ".pdf", ".ods"}))
        d.Show()
    })

    // Drag & Drop зона (кастомний віджет)
    dropper := widgets.NewFileDropper(func(paths []string) {
        for _, path := range paths {
            go func(p string) {
                txs, err := state.Registry.ParseFile(p)
                if err == nil {
                    state.Categorizer.CategorizeAll(txs)
                    state.Transactions = append(state.Transactions, txs...)
                    fileList.Refresh()
                }
            }(path)
        }
    })

    // Зведення після імпорту
    summary := widget.NewRichText()
    // Оновлюється після кожного імпорту

    return container.NewBorder(
        container.NewVBox(
            widget.NewLabel("Перетягніть файли або оберіть вручну"),
            dropper,
            addBtn,
        ),
        summary,
        nil, nil,
        fileList,
    )
}
```

### Таблиця транзакцій (`internal/ui/transactions_screen.go`)

```go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "bank-analyzer/internal/models"
)

func NewTransactionsScreen(state *AppState) fyne.CanvasObject {
    // Фільтри
    searchEntry := widget.NewEntry()
    searchEntry.SetPlaceHolder("Пошук по опису...")

    categorySelect := widget.NewSelect(state.Config.CategoryNames(), func(s string) {})
    categorySelect.PlaceHolder = "Всі категорії"

    typeSelect := widget.NewSelect([]string{"Всі", "Витрати", "Надходження"}, func(s string) {})

    // Таблиця
    table := widget.NewTable(
        func() (int, int) {
            return len(filtered(state, searchEntry.Text, categorySelect.Selected, typeSelect.Selected)) + 1, 7
        },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(id widget.TableCellID, cell fyne.CanvasObject) {
            label := cell.(*widget.Label)
            if id.Row == 0 {
                headers := []string{"Дата", "Тип", "Сума", "Категорія", "Опис", "Контрагент", "Джерело"}
                label.TextStyle = fyne.TextStyle{Bold: true}
                label.SetText(headers[id.Col])
                return
            }
            txs := filtered(state, searchEntry.Text, categorySelect.Selected, typeSelect.Selected)
            tx := txs[id.Row-1]
            switch id.Col {
            case 0: label.SetText(tx.Date.Format("02.01.2006"))
            case 1:
                if tx.Type == models.Debit {
                    label.SetText("↑ Витрата")
                } else {
                    label.SetText("↓ Надходження")
                }
            case 2: label.SetText(tx.Amount.StringFixed(2) + " " + tx.Currency)
            case 3: label.SetText(tx.Category)
            case 4: label.SetText(truncate(tx.Description, 40))
            case 5: label.SetText(tx.Counterparty)
            case 6: label.SetText(tx.BankSource)
            }
        },
    )

    // Подвійний клік — редагування категорії
    table.OnSelected = func(id widget.TableCellID) {
        if id.Row == 0 { return }
        txs := filtered(state, searchEntry.Text, categorySelect.Selected, typeSelect.Selected)
        tx := txs[id.Row-1]
        showCategoryEditor(tx, state)
    }

    // Колонки
    table.SetColumnWidth(0, 100)
    table.SetColumnWidth(1, 100)
    table.SetColumnWidth(2, 120)
    table.SetColumnWidth(3, 140)
    table.SetColumnWidth(4, 280)
    table.SetColumnWidth(5, 200)
    table.SetColumnWidth(6, 100)

    filterBar := container.NewHBox(searchEntry, categorySelect, typeSelect)

    // Статус-рядок
    statusLabel := widget.NewLabel("")
    updateStatus := func() {
        txs := filtered(state, searchEntry.Text, categorySelect.Selected, typeSelect.Selected)
        statusLabel.SetText(fmt.Sprintf("Знайдено: %d транзакцій", len(txs)))
    }
    searchEntry.OnChanged = func(s string) { table.Refresh(); updateStatus() }

    return container.NewBorder(filterBar, statusLabel, nil, nil, table)
}
```

### Екран звіту (`internal/ui/report_screen.go`)

```go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
    "bank-analyzer/internal/reports"
)

func NewReportScreen(state *AppState) fyne.CanvasObject {
    // Вибір формату
    formatSelect := widget.NewSelect([]string{"XLSX", "ODS"}, func(s string) {})
    formatSelect.SetSelected("XLSX")

    // Вибір шаблону (опціонально)
    templateLabel := widget.NewLabel("Шаблон: не обрано")
    templateBtn := widget.NewButton("Обрати шаблон...", func() {
        // file dialog для вибору template.xlsx
    })

    // Фільтр дат
    fromEntry := widget.NewEntry()
    fromEntry.SetPlaceHolder("від (дд.мм.рррр)")
    toEntry := widget.NewEntry()
    toEntry.SetPlaceHolder("до (дд.мм.рррр)")

    // Чекбокси секцій звіту
    includeSummary := widget.NewCheck("Підсумок", func(b bool) {})
    includeByCategory := widget.NewCheck("За категоріями", func(b bool) {})
    includeMonthly := widget.NewCheck("По місяцях", func(b bool) {})
    includeChart := widget.NewCheck("Графік", func(b bool) {})
    includeSummary.SetChecked(true)
    includeByCategory.SetChecked(true)
    includeMonthly.SetChecked(true)
    includeChart.SetChecked(true)

    // Прогрес та генерація
    progress := widget.NewProgressBar()
    progress.Hide()

    generateBtn := widget.NewButtonWithIcon("Згенерувати звіт", theme.DocumentSaveIcon(), func() {
        d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
            if writer == nil { return }
            progress.Show()
            go func() {
                reporter := &reports.XLSXReporter{}
                report := buildReport(state)
                err := reporter.Generate(report, writer.URI().Path())
                progress.Hide()
                if err != nil {
                    dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
                } else {
                    dialog.ShowInformation("Готово", "Звіт успішно збережено!", fyne.CurrentApp().Driver().AllWindows()[0])
                }
            }()
        }, fyne.CurrentApp().Driver().AllWindows()[0])
        d.SetFileName("звіт.xlsx")
        d.Show()
    })

    return container.NewVBox(
        widget.NewCard("Налаштування", "",
            container.NewVBox(
                container.NewHBox(widget.NewLabel("Формат:"), formatSelect),
                container.NewHBox(templateLabel, templateBtn),
                container.NewHBox(widget.NewLabel("Період:"), fromEntry, widget.NewLabel("—"), toEntry),
            ),
        ),
        widget.NewCard("Секції звіту", "",
            container.NewVBox(includeSummary, includeByCategory, includeMonthly, includeChart),
        ),
        progress,
        generateBtn,
    )
}
```

---

## go.mod залежності

```go
module bank-analyzer

go 1.22

require (
    fyne.io/fyne/v2 v2.5.0
    github.com/xuri/excelize/v2 v2.8.1      // XLSX читання/запис
    github.com/shopspring/decimal v1.4.0    // Точна арифметика грошей
    golang.org/x/text v0.16.0               // Кодування Windows-1251
    gopkg.in/yaml.v3 v3.0.1                 // Конфіг категорій
    github.com/pdfcpu/pdfcpu v0.7.0         // PDF парсинг
)
```

---

## Ключові рішення дизайну

### Точна арифметика
Використовуємо `shopspring/decimal` замість `float64` — критично для фінансових розрахунків (помилки округлення неприпустимі).

### Авто-детекція банку
Парсери реєструються з пріоритетом. `CanParse()` перевіряє не лише розширення, але й вміст файлу (заголовки CSV, структуру PDF) — користувачеві не потрібно вказувати банк вручну.

### Горутини для парсингу
Важкі операції (парсинг PDF, великих CSV) виконуються в горутинах щоб не блокувати UI. Прогрес передається через `chan`.

### Persistency налаштувань
Fyne зберігає конфіг через `app.Preferences()` (платформо-незалежно: `~/.config` на Linux, `AppData` на Windows, `Library` на macOS).

### Шаблони XLSX
При вказаному шаблоні `excelize.OpenFile()` завантажує існуючий файл, нові листи додаються до нього — корпоративні шаблони зі своїм стилем зберігаються.

---

## Roadmap / Розширення

| Фіча | Складність | Пріоритет |
|------|-----------|----------|
| Підтримка ODS (через go-ods) | Середня | Середній |
| ML-категоризація (на базі fastText) | Висока | Низький |
| Об'єднання дублікатів між банками | Середня | Високий |
| Monobank API (прямий імпорт) | Низька | Високий |
| Графіки всередині Fyne (canvas) | Висока | Середній |
| Мультивалютність (курси НБУ) | Середня | Середній |