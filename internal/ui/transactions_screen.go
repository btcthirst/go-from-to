package ui

import (
	"bank-analyzer/internal/models"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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
			case 0:
				label.SetText(tx.Date.Format("02.01.2006"))
			case 1:
				if tx.Type == models.Debit {
					label.SetText("↑ Витрата")
				} else {
					label.SetText("↓ Надходження")
				}
			case 2:
				label.SetText(tx.Amount.StringFixed(2) + " " + tx.Currency)
			case 3:
				label.SetText(tx.Category)
			case 4:
				label.SetText(truncate(tx.Description, 40))
			case 5:
				label.SetText(tx.Counterparty)
			case 6:
				label.SetText(tx.BankSource)
			}
		},
	)

	// Подвійний клік — редагування категорії
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
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

func filtered(state *AppState, search, category, txType string) []*models.Transaction {
	var result []*models.Transaction
	for _, tx := range state.Transactions {
		if search != "" && !containsIgnoreCase(tx.Description, search) && !containsIgnoreCase(tx.Counterparty, search) {
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func containsIgnoreCase(str, substr string) bool {
	return len(substr) == 0 || (len(str) >= len(substr) && (stringContainsFold(str, substr)))
}

func stringContainsFold(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		a, b := s[i], t[i]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}

func showCategoryEditor(tx *models.Transaction, state *AppState) {
	categorySelect := widget.NewSelect(state.Config.CategoryNames(), func(s string) {
		tx.Category = s
	})
	categorySelect.SetSelected(tx.Category)

	dialog.ShowCustom("Редагувати категорію", "Закрити", categorySelect, fyne.CurrentApp().Driver().AllWindows()[0])
}
