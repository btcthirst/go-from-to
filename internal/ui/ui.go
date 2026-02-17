// Package ui надає UI на основі Fyne для вибору файлів та обробки даних
package ui

import (
	"excel-parser/internal/app"
	"fmt"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// UIManager управляє UI застосунку
type UIManager struct {
	accrualFile string
	paymentFile string
	outputFile  string
}

// NewUIManager створює новий UI менеджер
func NewUIManager() *UIManager {
	return &UIManager{
		accrualFile: "",
		paymentFile: "",
		outputFile:  "",
	}
}

// Run запускає UI застосунку
func (um *UIManager) Run() {
	fyneAppInstance := fyneapp.New()
	w := fyneAppInstance.NewWindow("Excel/ODS File Processor")
	w.SetTitle("Обробник Excel/ODS файлів")

	// Лейбли для показу вибраних файлів
	accrualLabel := widget.NewLabel("Файл нарахувань: не вибрано")
	paymentLabel := widget.NewLabel("Файл виписок: не вибрано")
	outputLabel := widget.NewLabel("Файл для результатів: не вибрано")
	statusLabel := widget.NewLabel("")

	// Кнопка обробки (оголошуємо раніше)
	var processBtn *widget.Button

	// Кнопка вибору файлу нарахувань
	accrualBtn := widget.NewButton("Вибрати файл нарахувань (ODS)", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			um.accrualFile = reader.URI().Path()
			accrualLabel.SetText("Файл нарахувань: " + reader.URI().Name())
		}, w)
		fd.Show()
	})

	// Кнопка вибору файлу виписок
	paymentBtn := widget.NewButton("Вибрати файл виписок (XLSX)", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			um.paymentFile = reader.URI().Path()
			paymentLabel.SetText("Файл виписок: " + reader.URI().Name())
		}, w)
		fd.Show()
	})

	// Кнопка вибору місця для збереження
	outputBtn := widget.NewButton("Вибрати місце для результатів", func() {
		fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			um.outputFile = writer.URI().Path()
			outputLabel.SetText("Файл для результатів: " + writer.URI().Name())
		}, w)
		fd.SetFileName("summery.xlsx")
		fd.Show()
	})

	// Кнопка обробки
	processBtn = widget.NewButton("Обробити файли", func() {
		if um.accrualFile == "" || um.paymentFile == "" || um.outputFile == "" {
			dialog.ShowInformation("Помилка", "Будь ласка, виберіть усі потрібні файли", w)
			return
		}

		// Вимикаємо кнопку та показуємо статус
		processBtn.Disable()
		statusLabel.SetText("Обробка... будь ласка, зачекайте")

		// Запускаємо обробку в окремій горутині
		go func() {
			defer processBtn.Enable()

			// Викликаємо функцію обробки
			app.ProcessFiles(um.accrualFile, um.paymentFile, um.outputFile)

			statusLabel.SetText("Обробка завершена успішно!")
			dialog.ShowInformation("Успіх", fmt.Sprintf("Результати збережені у файл:\n%s", um.outputFile), w)
		}()
	})

	// Заголовок
	title := canvas.NewText("Обробник Excel/ODS файлів", nil)
	title.TextSize = 24
	title.TextStyle.Bold = true

	// Контейнер з інформацією про файли
	fileInfoBox := container.NewVBox(
		accrualLabel,
		paymentLabel,
		outputLabel,
	)

	// Контейнер з кнопками вибору файлів
	buttonsBox := container.NewVBox(
		accrualBtn,
		paymentBtn,
		outputBtn,
	)

	// Контейнер зі статусом
	statusBox := container.NewVBox(
		statusLabel,
	)

	// Основний контейнер
	mainBox := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel("Виберіть файли:"),
		buttonsBox,
		widget.NewSeparator(),
		fileInfoBox,
		widget.NewSeparator(),
		processBtn,
		statusBox,
	)

	w.SetContent(mainBox)
	w.ShowAndRun()
}
