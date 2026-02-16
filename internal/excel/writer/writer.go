// Package writer надає функціональність для запису результатів у файли Excel.
package writer

import (
	"excel-parser/internal/model"
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

// WriteResults записує результати у файл summery.xlsx
func WriteResults(results []model.ResultRecord) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())

	// Записуємо заголовок
	headers := []string{"№", "Дата", "Сума", "ПІБ", "Рахунок", "Контрагент"}
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return fmt.Errorf("помилка при формуванні назви комірки: %w", err)
		}
		f.SetCellStr(sheetName, cell, header)
	}

	// Записуємо дані
	for idx, result := range results {
		row := idx + 2
		
		cellNum, _ := excelize.CoordinatesToCellName(1, row)
		f.SetCellInt(sheetName, cellNum, idx+1)
		
		cellDate, _ := excelize.CoordinatesToCellName(2, row)
		f.SetCellStr(sheetName, cellDate, result.Date)
		
		cellSum, _ := excelize.CoordinatesToCellName(3, row)
		f.SetCellFloat(sheetName, cellSum, result.Sum, 2, 64)
		
		cellName, _ := excelize.CoordinatesToCellName(4, row)
		f.SetCellStr(sheetName, cellName, result.Name)
		
		cellAccount, _ := excelize.CoordinatesToCellName(5, row)
		f.SetCellStr(sheetName, cellAccount, result.Account)
		
		cellCounterparty, _ := excelize.CoordinatesToCellName(6, row)
		f.SetCellStr(sheetName, cellCounterparty, result.Counterparty)
	}

	// Зберігаємо файл
	if err := f.SaveAs("summery.xlsx"); err != nil {
		return fmt.Errorf("помилка при збереженні файлу: %w", err)
	}

	log.Println("Результати успішно записані у summery.xlsx")
	return nil
}
