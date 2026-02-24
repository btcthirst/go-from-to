// Package widgets provides custom UI components for the application, including the FileDroper widget for handling file drag-and-drop interactions.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func NewFileDropper(onDrop func(paths []fyne.URI)) *FileDropper {
	// Implement the FileDropper widget that detects file drag-and-drop events
	// and calls the onDrop callback with the list of file paths.
	f := &FileDropper{
		onDrop: onDrop,
	}
	f.ExtendBaseWidget(f)
	// Here you would set up the necessary event listeners for drag-and-drop
	// and call f.onDrop when files are dropped onto the widget.
	return f
}

type FileDropper struct {
	widget.BaseWidget
	onDrop func(paths []fyne.URI) // Callback function to handle dropped file paths
}

// Implement the necessary methods to make FileDropper a valid fyne.CanvasObject

func (f *FileDropper) CreateRenderer() fyne.WidgetRenderer {
	// Create a simple renderer for the FileDropper widget. This could be a rectangle with some text indicating that files can be dropped here.
	label := widget.NewLabel("Перетягніть файли сюди 📥")
	return widget.NewSimpleRenderer(label)
}

func (f *FileDropper) Dragged(event *fyne.DragEvent) {
	// Handle the drag event to provide visual feedback when files are dragged over the widget.
}

func (f *FileDropper) Dropped(event *fyne.DragEvent) {
	// Handle the drop event, extract file paths from the event, and call the onDrop callback.
	// This is where you would convert the event data into a list of fyne.URI and call f.onDrop.
}

func (f *FileDropper) Scrolled(event *fyne.ScrollEvent) {
	// Optionally handle scroll events if needed.
}

func (f *FileDropper) DropFiles(uris []fyne.URI) {
	if f.onDrop != nil {
		f.onDrop(uris)
	}
}
