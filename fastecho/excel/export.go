package excel

import (
	"github.com/xuri/excelize/v2"
)

// NewExcelDocument creates a new Excel file
func NewExcelDocument() (ExcelDocument, error) {
	file := excelize.NewFile()

	return ExcelDocument{File: file}, nil
}
