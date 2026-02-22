package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type CategoriesScreen struct {
	state *AppState
}

func NewCategoriesScreen(state *AppState) *CategoriesScreen {
	return &CategoriesScreen{
		state: state,
	}
}

func (s *CategoriesScreen) CreateUI() fyne.CanvasObject {
	// Тут можна реалізувати інтерфейс для перегляду та редагування категорій
	return widget.NewLabel("Категорії - ще в розробці")
}

func (s *CategoriesScreen) Hide() {
	// No-op for now, since there's no visible UI to hide in this screen
}

func (s *CategoriesScreen) Show() {
	// No-op for now, since there's no visible UI to show in this screen
}

func (s *CategoriesScreen) Refresh() {
	// No-op for now, since there's no dynamic content to refresh in this screen
}

func (s *CategoriesScreen) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

func (s *CategoriesScreen) Layout(size fyne.Size) {
	// No-op for now, since there's no dynamic layout needed in this screen
}

func (s *CategoriesScreen) Move(pos fyne.Position) {
	// No-op for now, since there's no dynamic positioning needed in this screen
}

func (s *CategoriesScreen) Size() fyne.Size {
	return s.MinSize()
}

func (s *CategoriesScreen) Position() fyne.Position {
	return fyne.NewPos(0, 0)
}

func (s *CategoriesScreen) Resize(size fyne.Size) {
	// No-op for now, since there's no dynamic resizing needed in this screen
}

func (s *CategoriesScreen) Visible() bool {
	return true
}
