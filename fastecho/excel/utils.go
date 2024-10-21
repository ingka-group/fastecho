package excel

import (
	"bytes"

	"github.com/xuri/excelize/v2"
)

// ExcelToMap unloads data from Excel spreadsheet as a map with sheet name as a key and rows as values
func ExcelToMap(excelFile *excelize.File) (map[string][][]string, error) {
	data := make(map[string][][]string)
	sheets := excelFile.GetSheetList()

	for _, sheet := range sheets {
		rows, err := excelFile.GetRows(sheet)
		if err != nil {
			return nil, err
		}

		data[sheet] = rows
	}

	return data, nil
}

// ReadFile reads an excel from a file
func ReadFile(filePath string) (*excelize.File, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// BytesToExcel parses a byte array into an excel file
func BytesToExcel(content []byte) (*excelize.File, error) {
	reader := bytes.NewReader(content)

	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	return file, nil
}
