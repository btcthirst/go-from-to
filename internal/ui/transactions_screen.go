package ui

import (
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/ui/widgets"
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Порядок колонок таблиці транзакцій.
const (
	colDate        = 0
	colType        = 1
	colAmount      = 2
	colProvider    = 3 // ← переміщено на 3-тю позицію
	colCategory    = 4
	colDescription = 5
	colCounterpart = 6
	colSource      = 7
	colCount       = 8
)

var tableHeaders = []string{
	"Дата", "Тип", "Сума", "Постачальник", "Категорія", "Опис", "Контрагент", "Джерело",
}

// Мінімальні ширини колонок. Користувач може розтягувати таблицю —
// Fyne дозволяє SetColumnWidth але не drag-resize нативно, тому для довгих
// клітинок використовуємо tooltip.
var columnWidths = []float32{100, 110, 110, 160, 110, 260, 180, 90}

// maxLen — після якої кількості рун показуємо "..." і вмикаємо tooltip.
const maxLen = 32

func NewTransactionsScreen(state *AppState, win fyne.Window) fyne.CanvasObject {
	// --- Фільтри ---
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Пошук по опису або контрагенту...")

	categorySelect := widget.NewSelect(state.Config.CategoryNames(), nil)
	categorySelect.PlaceHolder = "Всі категорії"

	typeSelect := widget.NewSelect([]string{"Всі", "Витрати", "Надходження"}, nil)
	typeSelect.SetSelected("Всі")

	statusLabel := widget.NewLabel("")

	// --- Кеш відфільтрованих транзакцій ---
	var cachedTxs []*models.Transaction

	refreshCache := func() {
		cachedTxs = filtered(state, searchEntry.Text, categorySelect.Selected, typeSelect.Selected)
		statusLabel.SetText(fmt.Sprintf("Знайдено: %d транзакцій", len(cachedTxs)))
	}
	refreshCache()

	// --- Таблиця ---
	// CreateItem повертає TooltipLabel для всіх клітинок — це дозволяє
	// показувати повний текст при наведенні на будь-яку обрізану клітинку.
	table := widget.NewTable(
		func() (rows, cols int) {
			return len(cachedTxs) + 1, colCount
		},
		func() fyne.CanvasObject {
			return widgets.NewTooltipLabel("", "")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widgets.TooltipLabel)

			// Рядок заголовків
			if id.Row == 0 {
				lbl.Style = fyne.TextStyle{Bold: true}
				lbl.SetTexts(tableHeaders[id.Col], "")
				return
			}

			if id.Row-1 >= len(cachedTxs) {
				lbl.SetTexts("", "")
				return
			}

			lbl.Style = fyne.TextStyle{}
			tx := cachedTxs[id.Row-1]

			switch id.Col {
			case colDate:
				v := tx.Date.Format("02.01.2006")
				lbl.SetTexts(v, "")

			case colType:
				if tx.Type == models.Debit {
					lbl.SetTexts("↑ Витрата", "")
				} else {
					lbl.SetTexts("↓ Надходження", "")
				}

			case colAmount:
				v := tx.Amount.StringFixed(2) + " " + tx.Currency
				lbl.SetTexts(v, "")

			case colProvider:
				lbl.SetTexts(truncate(tx.Provider, maxLen), tx.Provider)

			case colCategory:
				lbl.SetTexts(truncate(tx.Category, maxLen), tx.Category)

			case colDescription:
				lbl.SetTexts(truncate(tx.Description, maxLen), tx.Description)

			case colCounterpart:
				lbl.SetTexts(truncate(tx.Counterparty, maxLen), tx.Counterparty)

			case colSource:
				lbl.SetTexts(tx.BankSource, "")
			}
		},
	)

	for i, w := range columnWidths {
		table.SetColumnWidth(i, w)
	}

	// Клік на рядок — редагування категорії
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row-1 >= len(cachedTxs) {
			return
		}
		tx := cachedTxs[id.Row-1]
		showCategoryEditor(tx, state, win, func() {
			refreshCache()
			table.Refresh()
		})
	}

	// --- Фільтри ---
	applyFilters := func(_ string) {
		refreshCache()
		table.Refresh()
	}

	searchEntry.OnChanged = applyFilters
	categorySelect.OnChanged = applyFilters
	typeSelect.OnChanged = applyFilters

	resetBtn := widget.NewButton("Скинути", func() {
		searchEntry.SetText("")
		categorySelect.ClearSelected()
		typeSelect.SetSelected("Всі")
		refreshCache()
		table.Refresh()
	})

	filterBar := container.NewBorder(
		nil, nil, nil, resetBtn,
		container.NewGridWithColumns(3, searchEntry, categorySelect, typeSelect),
	)

	state.AddTransactionListener(func() {
		refreshCache()
		table.Refresh()
	})

	return container.NewBorder(filterBar, statusLabel, nil, nil, table)
}

// --- Фільтрація ---

func filtered(state *AppState, search, category, txType string) []*models.Transaction {
	txs := state.GetTransactions()
	result := make([]*models.Transaction, 0, len(txs))
	for _, tx := range txs {
		if search != "" &&
			!containsIgnoreCase(tx.Description, search) &&
			!containsIgnoreCase(tx.Counterparty, search) {
			continue
		}
		if category != "" && category != "Всі категорії" && tx.Category != category {
			continue
		}
		if txType == "Витрати" && tx.Type != models.Debit {
			continue
		}
		if txType == "Надходження" && tx.Type != models.Credit {
			continue
		}
		result = append(result, tx)
	}
	return result
}

// --- Допоміжні функції ---

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

func containsIgnoreCase(str, substr string) bool {
	if substr == "" {
		return true
	}
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// --- Діалог редагування транзакції ---

// showCategoryEditor відкриває діалог редагування категорії та постачальника.
func showCategoryEditor(tx *models.Transaction, state *AppState, win fyne.Window, onSave func()) {

	// --- Категорія ---
	selectedCategory := tx.Category
	categorySelect := widget.NewSelect(state.Config.CategoryNames(), func(s string) {
		selectedCategory = s
	})
	categorySelect.SetSelected(tx.Category)

	// --- Постачальник ---
	// Для категорії "Внески" — AutocompleteEntry з довідника + Select.
	// Для решти — звичайний Entry без автодоповнення.
	providerOptions := buildProviderOptions(state)

	var providerEntry interface {
		SetText(string)
		fyne.CanvasObject
	}
	var providerSuggestions fyne.CanvasObject

	if tx.Category == "Внески" {
		ac := widgets.NewAutocompleteEntry(providerOptions)
		ac.SetText(tx.Provider)
		ac.SetPlaceHolder("Постачальник...")
		providerEntry = ac
		providerSuggestions = ac.SuggestionsBox
	} else {
		e := widget.NewEntry()
		e.SetText(tx.Provider)
		e.SetPlaceHolder("Постачальник...")
		providerEntry = e
		providerSuggestions = nil
	}

	// Select з довідника — тільки для Внесків
	providerSelectRow := container.NewVBox()
	if tx.Category == "Внески" && len(providerOptions) > 0 {
		providerSelect := widget.NewSelect(providerOptions, func(s string) {
			providerEntry.SetText(s)
		})
		providerSelect.PlaceHolder = "Обрати з довідника..."
		providerSelectRow.Objects = []fyne.CanvasObject{
			widget.NewLabel("Або обрати з довідника:"),
			providerSelect,
		}
	}

	categorySelect.OnChanged = func(s string) {
		selectedCategory = s
	}

	manualCheck := widget.NewCheck("Зафіксувати вручну (не перезаписувати)", func(_ bool) {})
	manualCheck.SetChecked(tx.ProviderManual)

	formItems := []fyne.CanvasObject{
		widget.NewLabel("Категорія:"),
		categorySelect,
		widget.NewSeparator(),
		widget.NewLabel("Постачальник:"),
		providerEntry,
	}
	if providerSuggestions != nil {
		formItems = append(formItems, providerSuggestions)
	}
	formItems = append(formItems, providerSelectRow, manualCheck)
	form := container.NewVBox(formItems...)

	dialog.ShowCustomConfirm(
		"Редагувати транзакцію",
		"Зберегти",
		"Скасувати",
		form,
		func(confirm bool) {
			if !confirm {
				return
			}
			tx.Category = selectedCategory
			switch v := providerEntry.(type) {
			case *widgets.AutocompleteEntry:
				tx.Provider = strings.TrimSpace(v.Text)
			case *widget.Entry:
				tx.Provider = strings.TrimSpace(v.Text)
			}
			tx.ProviderManual = manualCheck.Checked
			if onSave != nil {
				onSave()
			}
		},
		win,
	)
}

// buildProviderOptions формує відсортований список "ПІБ(код)" з довідника.
func buildProviderOptions(state *AppState) []string {
	opts := make([]string, 0, len(state.Config.Providers.Providers))
	for code, name := range state.Config.Providers.Providers {
		opts = append(opts, fmt.Sprintf("%s(%s)", name, code))
	}
	sort.Strings(opts)
	return opts
}
