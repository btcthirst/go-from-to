package main

import (
	"excel-parser/internal/config"
	"excel-parser/internal/ui"
	"log"
)

func main() {
	// Завантажуємо конфігурацію
	if err := config.Load("config/document_config.yaml"); err != nil {
		log.Fatalf("Помилка завантаження конфігурації: %v\n", err)
	}

	// Запускаємо UI
	uiManager := ui.NewUIManager()
	uiManager.Run()
}
