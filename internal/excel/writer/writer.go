// Package writer надає функціональність для запису результатів у файли Excel/ODS.
package writer

import (
	"excel-parser/internal/model"
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

// WriteResults записує результати у файл (видбір формату за розширенням)
func WriteResults(results []model.ResultRecord, outputFile string) error {
	if strings.HasSuffix(strings.ToLower(outputFile), ".ods") {
		return WriteResultsODS(results, outputFile)
	}
	return WriteResultsXLSX(results, outputFile)
}

// WriteResultsXLSX записує результати у файл XLSX
func WriteResultsXLSX(results []model.ResultRecord, outputFile string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())

	// Записуємо заголовок
	headers := []string{
		"№п",
		"Постачальник",
		"Дата",
		"Дт рах.311 Сума",
		"Оборот по Дт",
		"313",
		"63",
		"641",
		"641.1",
		"651",
		"94",
		"Оборот по Кт",
	}
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
		
		cellSupplier, _ := excelize.CoordinatesToCellName(2, row)
		f.SetCellStr(sheetName, cellSupplier, result.Name)
		
		cellDate, _ := excelize.CoordinatesToCellName(3, row)
		f.SetCellStr(sheetName, cellDate, result.Date)
		
		cellSum, _ := excelize.CoordinatesToCellName(4, row)
		f.SetCellFloat(sheetName, cellSum, result.Sum, 2, 64)
		
		// Дані для рахунків (313, 63, 641, 641.1, 651, 94) - поки користуємо Account як індикатор
		cellAccount, _ := excelize.CoordinatesToCellName(5, row)
		f.SetCellStr(sheetName, cellAccount, result.Account)
		
		// Контрагент у 12 колонку (Оборот по Кт)
		cellCounterparty, _ := excelize.CoordinatesToCellName(12, row)
		f.SetCellStr(sheetName, cellCounterparty, result.Counterparty)
	}

	// Зберігаємо файл
	if err := f.SaveAs(outputFile); err != nil {
		return fmt.Errorf("помилка при збереженні файлу: %w", err)
	}

	log.Printf("Результати успішно записані у %s\n", outputFile)
	return nil
}
