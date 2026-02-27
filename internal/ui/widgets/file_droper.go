// Package widgets provides custom UI components for the application,
// including the FileDropper widget for handling file drag-and-drop interactions.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// FileDropper — віджет що приймає файли через drag-and-drop.
type FileDropper struct {
	widget.BaseWidget
	onDrop  func(uris []fyne.URI)
	hovered bool
}

// Перевіряємо лише Hoverable. desktop.FileDropper не існує у Fyne API.
var _ desktop.Hoverable = (*FileDropper)(nil)

func NewFileDropper(onDrop func(uris []fyne.URI)) *FileDropper {
	f := &FileDropper{onDrop: onDrop}
	f.ExtendBaseWidget(f)
	return f
}

// DroppedFiles викликається з екрану (Window OnDropped), коли файли скинуті в межі віджета.
func (f *FileDropper) DroppedFiles(uris []fyne.URI) {
	if f.onDrop != nil {
		f.onDrop(uris)
	}
	// Знімаємо стан наведення, щоб віджет не "залипав" після відпускання файлу
	f.hovered = false
	f.Refresh()
}

// MouseIn /MouseOut — підсвітка при наведенні.
func (f *FileDropper) MouseIn(_ *desktop.MouseEvent) {
	f.hovered = true
	f.Refresh()
}

func (f *FileDropper) MouseMoved(_ *desktop.MouseEvent) {}

func (f *FileDropper) MouseOut() {
	f.hovered = false
	f.Refresh()
}

// --- Renderer ---

func (f *FileDropper) CreateRenderer() fyne.WidgetRenderer {
	border := canvas.NewRectangle(theme.InputBackgroundColor())
	border.StrokeWidth = 2
	border.StrokeColor = theme.InputBorderColor()
	border.CornerRadius = theme.SelectionRadiusSize()

	label := widget.NewLabel("Перетягніть файли сюди 📥")
	label.Alignment = fyne.TextAlignCenter

	return &fileDropperRenderer{owner: f, border: border, label: label}
}

type fileDropperRenderer struct {
	owner  *FileDropper
	border *canvas.Rectangle
	label  *widget.Label
}

func (r *fileDropperRenderer) Layout(size fyne.Size) {
	r.border.Resize(size)
	r.label.Move(fyne.NewPos(0, (size.Height-r.label.MinSize().Height)/2))
	r.label.Resize(fyne.NewSize(size.Width, r.label.MinSize().Height))
}

func (r *fileDropperRenderer) MinSize() fyne.Size {
	// Динамічний розмір, щоб текст ніколи не обрізався
	lblMin := r.label.MinSize()
	padding := theme.Padding() * 4
	return fyne.NewSize(fyne.Max(200, lblMin.Width+padding), fyne.Max(80, lblMin.Height+padding))
}

func (r *fileDropperRenderer) Refresh() {
	if r.owner.hovered {
		r.border.StrokeColor = theme.PrimaryColor()
		r.border.FillColor = theme.HoverColor()
	} else {
		r.border.StrokeColor = theme.InputBorderColor()
		r.border.FillColor = theme.InputBackgroundColor()
	}
	r.border.CornerRadius = theme.SelectionRadiusSize()
	canvas.Refresh(r.border)
}

func (r *fileDropperRenderer) Destroy() {}

func (r *fileDropperRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.border, r.label}
}
