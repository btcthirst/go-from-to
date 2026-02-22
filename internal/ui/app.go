// Package ui provides the user interface components for the application.
package ui

import (
	"bank-analyzer/internal/categoryzer"
	"bank-analyzer/internal/config"
	"bank-analyzer/internal/models"
	"bank-analyzer/internal/parsers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	Transactions []*models.Transaction
	Registry     *parsers.Registry
	Categorizer  *categoryzer.RulesCategorizer
	Config       *config.Config
	loadedFiles  []string
}

func Run() {
	config := config.Load()
	registry := parsers.NewRegistry()
	categorizer, err := categoryzer.NewRulesCategorizer(config.ConfigPath)
	if err != nil {
		panic("Failed to load categorization rules: " + err.Error())
	}
	state := &AppState{
		Transactions: []*models.Transaction{},
		Registry:     registry,
		Categorizer:  categorizer,
		Config:       config,
	}
	a := app.NewWithID("ua.bankanalyzer")
	w := a.NewWindow("Bank Analyzer")

	w.Resize(fyne.NewSize(800, 600))

	nav := widget.NewList(
		func() int { return 4 },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			labels := []string{"📂 Імпорт", "📋 Транзакції", "🏷️ Категорії", "📊 Звіт"}
			o.(*widget.Label).SetText(labels[i])
		},
	)

	content := container.NewStack()
	screens := []fyne.CanvasObject{
		NewImportScreen(state),
		NewTransactionsScreen(state),
		NewCategoriesScreen(state),
		NewReportScreen(state),
	}

	nav.OnSelected = func(id widget.ListItemID) {
		content.Objects = []fyne.CanvasObject{screens[id]}
		content.Refresh()
	}

	nav.Select(0) // Select the first item by default
	split := container.NewHSplit(nav, content)
	split.SetOffset(0.2)

	w.SetContent(split)
	w.ShowAndRun()
}
