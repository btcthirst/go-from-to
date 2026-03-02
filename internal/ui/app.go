// Package ui provides the user interface components for the application.
package ui

import (
	"sync"

	"bank-analyzer/internal/categorizer"
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
	mu           sync.RWMutex
	Transactions []*models.Transaction
	Registry     *parsers.Registry
	Categorizer  *categorizer.RulesCategorizer
	Resolver     *categorizer.ProviderResolver
	Config       *config.Config

	txListeners []func()
}

func (s *AppState) AddTransactionListener(fn func()) {
	s.txListeners = append(s.txListeners, fn)
}

func (s *AppState) AppendTransactions(txs []*models.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Transactions = append(s.Transactions, txs...)
}

func (s *AppState) GetTransactions() []*models.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Transactions
}

func (s *AppState) ClearTransactions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Transactions = []*models.Transaction{}
}

// NotifyTransactionsChanged безпечно викликає колбек, якщо він встановлений.
func (s *AppState) NotifyTransactionsChanged() {
	for _, fn := range s.txListeners {
		fn()
	}
}

func Run() {
	cfg := config.Load()
	registry := parsers.NewRegistry(cfg.Mappings)

	cat := categorizer.NewRulesCategorizer(cfg.Categories)
	resolv := categorizer.NewProviderResolver(cfg.Providers)

	state := &AppState{
		mu:           sync.RWMutex{},
		Transactions: make([]*models.Transaction, 0),
		Registry:     registry,
		Categorizer:  cat,
		Resolver:     resolv,
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
