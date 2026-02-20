package main

import (
	"fmt"
	"log"

	"excel-parser/internal/config"
	"excel-parser/internal/excel/reader"
)

func main() {
	// Завантажуємо конфігурацію
	if err := config.Load("config/document_config.yaml"); err != nil {
		log.Fatalf("Помилка завантаження конфігурації: %v\n", err)
	}

	fmt.Println("=== АНАЛІЗ ФАЙЛУ нарахування26.ods ===")
	doc, err := reader.OpenODS("нарахування26.ods")
	if err != nil {
		fmt.Printf("Помилка: %v\n", err)
		return
	}

	sheetNames := doc.GetAllSheetNames()
	if len(sheetNames) == 0 {
		fmt.Println("Не знайдено листів у документі")
		return
	}

	// Get first sheet
	sheetData, err := doc.GetSheet(sheetNames[0])
	if err != nil {
		fmt.Printf("Помилка отримання листа: %v\n", err)
		return
	}

	// Display first 15 rows
	for i, row := range sheetData {
		if i >= 15 {
			break
		}
		fmt.Printf("Рядок %d: %v\n", i+1, row)
	}
}
