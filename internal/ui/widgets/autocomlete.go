package widgets

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AutocompleteEntry — Entry з вбудованим списком підказок.
//
// Використання в layout:
//
//	entry := widgets.NewAutocompleteEntry(opts)
//	form := container.NewVBox(entry, entry.SuggestionsBox)
type AutocompleteEntry struct {
	widget.Entry

	AllOptions []string
	OnSelected func(s string)

	// SuggestionsBox — додайте одразу після Entry у layout.
	SuggestionsBox fyne.CanvasObject

	list     *widget.List
	sizer    *listSizer
	filtered []string
}

func NewAutocompleteEntry(options []string) *AutocompleteEntry {
	e := &AutocompleteEntry{
		AllOptions: options,
		filtered:   []string{},
	}
	e.ExtendBaseWidget(e)

	e.list = widget.NewList(
		func() int { return len(e.filtered) },
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(e.filtered) {
				obj.(*widget.Label).SetText(e.filtered[id])
			}
		},
	)
	e.list.OnSelected = func(id widget.ListItemID) {
		if id >= len(e.filtered) {
			return
		}
		chosen := e.filtered[id]
		e.filtered = e.filtered[:0]
		e.list.UnselectAll()
		e.list.Refresh()
		e.SuggestionsBox.Hide()
		e.Entry.SetText(chosen)
		if e.OnSelected != nil {
			e.OnSelected(chosen)
		}
	}

	e.sizer = &listSizer{list: e.list}
	e.SuggestionsBox = container.NewStack(e.sizer)
	e.SuggestionsBox.Hide()

	return e
}

// --- Entry overrides ---

func (e *AutocompleteEntry) TypedRune(r rune) {
	e.Entry.TypedRune(r)
	e.updateSuggestions()
}

func (e *AutocompleteEntry) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyEscape:
		e.clearSuggestions()
	case fyne.KeyBackspace, fyne.KeyDelete:
		e.Entry.TypedKey(ev)
		e.updateSuggestions()
	default:
		e.Entry.TypedKey(ev)
	}
}

// --- Internal ---

func (e *AutocompleteEntry) updateSuggestions() {
	query := strings.ToLower(strings.TrimSpace(e.Text))

	e.filtered = e.filtered[:0]
	if query != "" {
		for _, opt := range e.AllOptions {
			if strings.Contains(strings.ToLower(opt), query) {
				e.filtered = append(e.filtered, opt)
			}
		}
	}

	if len(e.filtered) == 0 {
		e.SuggestionsBox.Hide()
		return
	}

	rows := len(e.filtered)
	if rows > 5 {
		rows = 5
	}
	rowH := e.Entry.MinSize().Height + theme.Padding()*2
	e.sizer.height = float32(rows) * rowH
	e.list.Refresh()
	e.SuggestionsBox.Show()
	e.SuggestionsBox.Refresh()
}

func (e *AutocompleteEntry) clearSuggestions() {
	e.filtered = e.filtered[:0]
	e.SuggestionsBox.Hide()
}

// --- listSizer — обгортка що задає мінімальну висоту для List ---

type listSizer struct {
	widget.BaseWidget
	list   *widget.List
	height float32
}

func (s *listSizer) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)
	return widget.NewSimpleRenderer(s.list)
}

func (s *listSizer) MinSize() fyne.Size {
	base := s.list.MinSize()
	if s.height > 0 {
		return fyne.NewSize(base.Width, s.height)
	}
	return base
}
