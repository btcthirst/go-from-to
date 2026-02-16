package main

import (
	"fmt"

	"github.com/unidoc/unioffice/spreadsheet"
)

func main() {
	fmt.Println("=== АНАЛІЗ ФАЙЛУ нарахування26.ods ===")
	doc, err := spreadsheet.Open("нарахування26.ods")
	if err != nil {
		fmt.Printf("Помилка: %v\n", err)
		return
	}
	defer doc.Close()

	sheets := doc.Sheets()
	if len(sheets) == 0 {
		fmt.Println("Не знайдено листів у документі")
		return
	}

	sheet := sheets[0]
	rows := sheet.Rows()

	for i, row := range rows {
		if i >= 15 {
			break
		}
		var cols []string
		for _, cell := range row.Cells() {
			cols = append(cols, cell.GetString())
		}
		fmt.Printf("Рядок %d: %v\n", i+1, cols)
	}
}
