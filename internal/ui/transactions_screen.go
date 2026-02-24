package ui

import (
	"bank-analyzer/internal/models"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	colDate        = 0
	colType        = 1
	colAmount      = 2
	colCategory    = 3
	colDescription = 4
	colCounterpart = 5
	colSource      = 6
	colCount       = 7
)

var tableHeaders = []string{"Дата", "Тип", "Сума", "Категорія", "Опис", "Контрагент", "Джерело"}

var columnWidths = []float32{100, 110, 130, 150, 280, 200, 110}

func NewTransactionsScreen(state *AppState) fyne.CanvasObject {
	// --- Фільтри ---
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Пошук по опису або контрагенту...")

	categorySelect := widget.NewSelect(state.Config.CategoryNames(), nil)
	categorySelect.PlaceHolder = "Всі категорії"

	typeSelect := widget.NewSelect([]string{"Всі", "Витрати", "Надходження"}, nil)
	typeSelect.SetSelected("Всі")

	// Статус-рядок
	statusLabel := widget.NewLabel("")

	// --- Кеш відфільтрованих транзакцій ---
	// Єдине місце, де filtered() викликається — в refreshCache.
	// Таблиця читає тільки з cachedTxs, не викликаючи filtered() на кожну клітинку.
	var cachedTxs []*models.Transaction

	refreshCache := func() {
		cachedTxs = filtered(state, searchEntry.Text, categorySelect.Selected, typeSelect.Selected)
		statusLabel.SetText(fmt.Sprintf("Знайдено: %d транзакцій", len(cachedTxs)))
	}

	// Ініціалізуємо кеш одразу
	refreshCache()

	// --- Таблиця ---
	table := widget.NewTable(
		func() (rows, cols int) {
			return len(cachedTxs) + 1, colCount // +1 — рядок заголовків
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			// Рядок заголовків
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(tableHeaders[id.Col])
				return
			}

			// Захист від виходу за межі під час рефрешу
			if id.Row-1 >= len(cachedTxs) {
				label.SetText("")
				return
			}

			tx := cachedTxs[id.Row-1]
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case colDate:
				label.SetText(tx.Date.Format("02.01.2006"))
			case colType:
				if tx.Type == models.Debit {
					label.SetText("↑ Витрата")
				} else {
					label.SetText("↓ Надходження")
				}
			case colAmount:
				label.SetText(tx.Amount.StringFixed(2) + " " + tx.Currency)
			case colCategory:
				label.SetText(tx.Category)
			case colDescription:
				label.SetText(truncate(tx.Description, 40))
			case colCounterpart:
				label.SetText(tx.Counterparty)
			case colSource:
				label.SetText(tx.BankSource)
			}
		},
	)

	// Ширини колонок
	for i, w := range columnWidths {
		table.SetColumnWidth(i, w)
	}

	// Клік на рядок — редагування категорії
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row-1 >= len(cachedTxs) {
			return
		}
		tx := cachedTxs[id.Row-1]
		showCategoryEditor(tx, state, func() {
			// Після зміни категорії — оновити кеш і таблицю
			refreshCache()
			table.Refresh()
		})
	}

	// --- Підключення фільтрів ---
	applyFilters := func(_ string) {
		refreshCache()
		table.Refresh()
	}

	searchEntry.OnChanged = applyFilters
	categorySelect.OnChanged = applyFilters
	typeSelect.OnChanged = applyFilters

	// --- Кнопка скидання фільтрів ---
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

	state.OnTransactionsChanged = func() {
		refreshCache()
		table.Refresh()
	}

	return container.NewBorder(filterBar, statusLabel, nil, nil, table)
}

// Refresh дозволяє зовнішньому коду (наприклад, після імпорту) оновити екран.
// Використовується через type assertion: screen.(*transactionsScreen) якщо потрібно,
// або простіше — зберігати refreshCache як поле AppState.
// Поки достатньо викликати table.Refresh() через fyne.Do після імпорту.

// --- Фільтрація ---

func filtered(state *AppState, search, category, txType string) []*models.Transaction {
	result := make([]*models.Transaction, 0, len(state.Transactions))

	for _, tx := range state.Transactions {
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

// truncate скорочує рядок до max рун (не байтів), додаючи "..." в кінці.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// containsIgnoreCase перевіряє наявність підрядка без урахування регістру,
// коректно працює з кирилицею через strings.ToLower (Unicode-aware).
func containsIgnoreCase(str, substr string) bool {
	if substr == "" {
		return true
	}
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// --- Діалог редагування категорії ---

// showCategoryEditor відкриває діалог зміни категорії транзакції.
// onSave викликається після підтвердження зміни.
func showCategoryEditor(tx *models.Transaction, state *AppState, onSave func()) {
	selected := tx.Category

	categorySelect := widget.NewSelect(state.Config.CategoryNames(), func(s string) {
		selected = s
	})
	categorySelect.SetSelected(tx.Category)

	win := fyne.CurrentApp().Driver().AllWindows()[0]

	dialog.ShowCustomConfirm(
		"Редагувати категорію",
		"Зберегти",
		"Скасувати",
		categorySelect,
		func(confirm bool) {
			if !confirm {
				return
			}
			tx.Category = selected
			if onSave != nil {
				onSave()
			}
		},
		win,
	)
}
