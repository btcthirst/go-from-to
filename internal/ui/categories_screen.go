package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewCategoriesScreen повертає екран управління категоріями та ключовими словами.
func NewCategoriesScreen(state *AppState) fyne.CanvasObject {
	win := fyne.CurrentApp().Driver().AllWindows()[0]

	// --- Стан екрану ---
	categories := state.Config.CategoryNames()
	var selectedCategory string

	// --- Ліва панель: список категорій ---
	categoryList := widget.NewList(
		func() int { return len(categories) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.GridIcon()),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(categories) {
				return
			}
			box := item.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			label.SetText(categories[id])

			// Підрахунок транзакцій у категорії
			count := countTxInCategory(state, categories[id])
			if count > 0 {
				label.SetText(fmt.Sprintf("%s (%d)", categories[id], count))
			}
		},
	)

	// --- Права панель: деталі категорії ---
	categoryNameLabel := widget.NewLabelWithStyle(
		"Оберіть категорію",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	txCountLabel := widget.NewLabel("")
	txCountLabel.Importance = widget.LowImportance

	// Список ключових слів
	keywordList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewSeparator(),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {},
	)

	// Поле додавання ключового слова
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("Нове ключове слово...")

	addKeywordBtn := widget.NewButtonWithIcon("Додати", theme.ContentAddIcon(), nil)
	addKeywordBtn.Disable()

	// Активуємо кнопку тільки якщо є текст і обрана категорія
	keywordEntry.OnChanged = func(s string) {
		if strings.TrimSpace(s) != "" && selectedCategory != "" {
			addKeywordBtn.Enable()
		} else {
			addKeywordBtn.Disable()
		}
	}

	// --- Функція оновлення правої панелі ---
	var currentKeywords []string
	var refreshDetail func(category string)
	refreshDetail = func(category string) {
		selectedCategory = category
		currentKeywords = state.Config.KeywordsForCategory(category)

		categoryNameLabel.SetText(category)

		count := countTxInCategory(state, category)
		if count > 0 {
			txCountLabel.SetText(fmt.Sprintf("%d транзакцій у цій категорії", count))
		} else {
			txCountLabel.SetText("Транзакцій немає")
		}

		keywordList.Length = func() int { return len(currentKeywords) }
		keywordList.CreateItem = func() fyne.CanvasObject {
			kw := widget.NewLabel("")
			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			delBtn.Importance = widget.DangerImportance
			return container.NewBorder(nil, nil, nil, delBtn, kw)
		}
		keywordList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(currentKeywords) {
				return
			}
			kw := currentKeywords[id]
			box := item.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			delBtn := box.Objects[1].(*widget.Button)

			label.SetText(kw)
			delBtn.OnTapped = func() {
				dialog.ShowConfirm(
					"Видалити ключове слово",
					fmt.Sprintf("Видалити «%s» з категорії «%s»?", kw, selectedCategory),
					func(ok bool) {
						if !ok {
							return
						}
						state.Config.RemoveKeyword(selectedCategory, kw)
						refreshDetail(selectedCategory)
						categoryList.Refresh()
					},
					win,
				)
			}
		}
		keywordList.Refresh()

		// Скидаємо поле вводу
		keywordEntry.SetText("")
		addKeywordBtn.Disable()
	}

	// --- Обробник вибору категорії ---
	categoryList.OnSelected = func(id widget.ListItemID) {
		if id >= len(categories) {
			return
		}
		refreshDetail(categories[id])
	}

	// --- Додавання ключового слова ---
	addKeywordBtn.OnTapped = func() {
		kw := strings.TrimSpace(keywordEntry.Text)
		if kw == "" || selectedCategory == "" {
			return
		}
		state.Config.AddKeyword(selectedCategory, kw)
		refreshDetail(selectedCategory)
		categoryList.Refresh()
	}

	// Додавання по Enter
	keywordEntry.OnSubmitted = func(s string) {
		addKeywordBtn.OnTapped()
	}

	// --- Додавання нової категорії ---
	addCategoryBtn := widget.NewButtonWithIcon("Нова категорія", theme.ContentAddIcon(), func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Назва категорії...")

		dialog.ShowCustomConfirm(
			"Додати категорію",
			"Додати",
			"Скасувати",
			nameEntry,
			func(ok bool) {
				if !ok {
					return
				}
				name := strings.TrimSpace(nameEntry.Text)
				if name == "" {
					return
				}
				// Перевірка на дублікат
				for _, c := range categories {
					if strings.EqualFold(c, name) {
						dialog.ShowInformation("Вже існує", fmt.Sprintf("Категорія «%s» вже є у списку.", name), win)
						return
					}
				}
				state.Config.AddCategory(name)
				categories = state.Config.CategoryNames()
				categoryList.Refresh()
			},
			win,
		)
	})

	// --- Видалення категорії ---
	deleteCategoryBtn := widget.NewButtonWithIcon("Видалити", theme.DeleteIcon(), func() {
		if selectedCategory == "" {
			return
		}
		dialog.ShowConfirm(
			"Видалити категорію",
			fmt.Sprintf("Видалити категорію «%s»?\nТранзакції з цією категорією стануть «Uncategorized».", selectedCategory),
			func(ok bool) {
				if !ok {
					return
				}
				// Переводимо транзакції у Uncategorized
				for _, tx := range state.Transactions {
					if tx.Category == selectedCategory {
						tx.Category = "Uncategorized"
					}
				}
				state.Config.RemoveCategory(selectedCategory)
				categories = state.Config.CategoryNames()
				selectedCategory = ""
				categoryNameLabel.SetText("Оберіть категорію")
				txCountLabel.SetText("")
				currentKeywords = nil
				keywordList.Refresh()
				categoryList.UnselectAll()
				categoryList.Refresh()
			},
			win,
		)
	})
	deleteCategoryBtn.Importance = widget.DangerImportance

	// --- Layout ---

	// Ліва панель
	leftToolbar := container.NewHBox(addCategoryBtn)
	leftPanel := container.NewBorder(leftToolbar, nil, nil, nil, categoryList)

	// Права панель — деталі категорії
	keywordInputRow := container.NewBorder(nil, nil, nil, addKeywordBtn, keywordEntry)

	detailHeader := container.NewVBox(
		categoryNameLabel,
		txCountLabel,
		widget.NewSeparator(),
		widget.NewLabel("Ключові слова:"),
	)

	detailFooter := container.NewVBox(
		widget.NewSeparator(),
		keywordInputRow,
		container.NewHBox(deleteCategoryBtn),
	)

	rightPanel := container.NewBorder(detailHeader, detailFooter, nil, nil, keywordList)

	// Розділювач між панелями
	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.35)

	return split
}

// countTxInCategory повертає кількість транзакцій з даною категорією.
func countTxInCategory(state *AppState, category string) int {
	count := 0
	for _, tx := range state.GetTransactions() {
		if tx.Category == category {
			count++
		}
	}
	return count
}
