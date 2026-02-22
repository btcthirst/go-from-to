// Package widgets provides custom UI components for the application, including the FileDroper widget for handling file drag-and-drop interactions.
package widgets

import "fyne.io/fyne/v2"

func NewFileDropper(onDrop func(paths []string)) *FileDropper {
	// Implement the FileDropper widget that detects file drag-and-drop events
	// and calls the onDrop callback with the list of file paths.
	return &FileDropper{
		onDrop: onDrop,
	}
}

type FileDropper struct {
	onDrop func(paths []string)
}

// Implement the necessary methods to make FileDropper a valid fyne.CanvasObject

func (f *FileDropper) MinSize() fyne.Size {
	return fyne.NewSize(200, 200)
}

func (f *FileDropper) Layout(size fyne.Size) {
	// No layout needed for a simple dropper widget
}

func (f *FileDropper) Refresh() {
	// No refresh needed for a simple dropper widget
}

func (f *FileDropper) Show() {
	// No show needed for a simple dropper widget
}

func (f *FileDropper) Hide() {
	// No hide needed for a simple dropper widget
}

func (f *FileDropper) Move(pos fyne.Position) {
	// No move needed for a simple dropper widget
}

func (f *FileDropper) Size() fyne.Size {
	return f.MinSize()
}

func (f *FileDropper) Position() fyne.Position {
	return fyne.NewPos(0, 0)
}

func (f *FileDropper) Visible() bool {
	return true
}

func (f *FileDropper) Resize(size fyne.Size) {
	// No resize needed for a simple dropper widget
}
