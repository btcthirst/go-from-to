// Package ui provides the user interface components for the application.
package ui

import (
	"log"

	"bank-analyzer/internal/categoryzer"
	"bank-analyzer/internal/config"
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/parsers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AppState — спільний стан застосунку, доступний усім екранам.
type AppState struct {
	Transactions []*models.Transaction
	Registry     *parsers.Registry
	Categorizer  *categoryzer.RulesCategorizer
	Config       *config.Config

	// OnTransactionsChanged викликається після будь-якої зміни Transactions
	// (імпорт, очищення). Екрани підписуються на цей колбек для оновлення.
	OnTransactionsChanged func()
}

// NotifyTransactionsChanged безпечно викликає колбек, якщо він встановлений.
func (s *AppState) NotifyTransactionsChanged() {
	if s.OnTransactionsChanged != nil {
		s.OnTransactionsChanged()
	}
}

func Run() {
	cfg := config.Load()
	registry := parsers.NewRegistry(cfg.Mappings)

	cat, err := categoryzer.NewRulesCategorizer(cfg.ConfigPath)
	if err != nil {
		// Не падаємо — запускаємось з порожнім категоризатором і логуємо проблему.
		// Користувач побачить транзакції без категорій, але застосунок працює.
		log.Printf("WARNING: не вдалося завантажити rules категоризатора (%s): %v. Запуск без категоризації.", cfg.ConfigPath, err)
		cat = categoryzer.NewEmptyCategorizer()
	}

	state := &AppState{
		Transactions: make([]*models.Transaction, 0),
		Registry:     registry,
		Categorizer:  cat,
		Config:       cfg,
	}

	a := app.NewWithID("ua.bankanalyzer")
	w := a.NewWindow("Bank Analyzer")
	w.Resize(fyne.NewSize(1280, 750))
	w.SetMaster()

	// --- Екрани ---
	// NewTransactionsScreen сам підписується на state.OnTransactionsChanged.
	screens := []fyne.CanvasObject{
		NewImportScreen(state),
		NewTransactionsScreen(state),
		NewCategoriesScreen(state),
		NewReportScreen(state),
	}

	// --- Навігація ---
	type navItem struct {
		label string
		icon  fyne.Resource
	}

	navItems := []navItem{
		{"Імпорт", theme.FolderOpenIcon()},
		{"Транзакції", theme.ListIcon()},
		{"Категорії", theme.GridIcon()},
		{"Звіт", theme.DocumentIcon()},
	}

	content := container.NewStack(screens[0])

	var navButtons []*widget.Button
	selectScreen := func(idx int) {
		for i, btn := range navButtons {
			if i == idx {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
			}
			btn.Refresh()
		}
		content.Objects = []fyne.CanvasObject{screens[idx]}
		content.Refresh()
	}

	for i, item := range navItems {
		i, item := i, item // capture
		btn := widget.NewButtonWithIcon(item.label, item.icon, func() {
			selectScreen(i)
		})
		btn.Importance = widget.MediumImportance
		navButtons = append(navButtons, btn)
	}

	// Підсвічуємо перший екран одразу
	navButtons[0].Importance = widget.HighImportance

	navPanel := container.NewVBox()
	for _, btn := range navButtons {
		navPanel.Add(btn)
	}
	navPanel.Add(widget.NewSeparator())

	// Версія / назва внизу панелі навігації
	appLabel := widget.NewLabelWithStyle("Bank Analyzer", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	appLabel.Importance = widget.LowImportance

	navContainer := container.NewBorder(nil, appLabel, nil, nil, navPanel)

	split := container.NewHSplit(navContainer, content)
	split.SetOffset(0.18)

	w.SetContent(split)
	w.ShowAndRun()
}
