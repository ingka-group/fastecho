package excel

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type ExcelStyle uint

// ExcelSheetColumn width sizes
const (
	S   = 10
	M   = 15
	L   = 20
	XL  = 35
	XXL = 40
)

const (
	ExcelStyleBold ExcelStyle = iota
)

type ExcelDocument struct {
	File   *excelize.File
	Sheets []ExcelSheet
}

// AddSheet adds a new sheet to the Excel document
func (f *ExcelDocument) AddSheet(name string) (*ExcelSheet, error) {
	// create new sheet in the document
	_, err := f.File.NewSheet(name)
	if err != nil {
		return nil, err
	}

	// create a writer for the new sheet
	writer, err := f.File.NewStreamWriter(name)
	if err != nil {
		return nil, err
	}

	sheet := ExcelSheet{
		Name:       name,
		CurrentRow: 1,
		Writer:     writer,
	}

	f.Sheets = append(f.Sheets, sheet)

	return &sheet, nil
}

// WriteToBuffer writes the data to a bytes.Buffer
func (f *ExcelDocument) WriteToBuffer() (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	for _, s := range f.Sheets {
		err := s.Writer.Flush()
		if err != nil {
			return nil, err
		}
	}

	bs, err := f.File.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	_, err = buffer.Write(bs.Bytes())
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

// AddStyle adds a new style to the Excel document
func (f *ExcelDocument) AddStyle(st ExcelStyle) (int, error) {
	if st == ExcelStyleBold {
		return f.File.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	}

	return 0, fmt.Errorf("excel style is not supported: %d", st)
}

type ExcelSheetColumn struct {
	Name  string
	Width float64
}

type ExcelSheet struct {
	Name         string
	CurrentRow   int
	CurrentStyle int

	Writer *excelize.StreamWriter
}

// SetColumns sets the number of columns along with the width to the Excel file
func (e *ExcelSheet) SetColumns(cols []ExcelSheetColumn) error {
	// Set width iterate cols and create structure for SetRow
	var wCols []interface{}
	for i := range cols {
		pos := i + 1

		err := e.Writer.SetColWidth(pos, pos, cols[i].Width)
		if err != nil {
			return err
		}

		wCols = append(wCols, cols[i].Name)
	}

	err := e.AddRow(wCols)
	if err != nil {
		return err
	}

	return nil
}

// AddRow adds a row to the Excel file
func (e *ExcelSheet) AddRow(elements []interface{}) error {
	cell, err := excelize.CoordinatesToCellName(1, e.CurrentRow)
	if err != nil {
		return err
	}

	opts := excelize.RowOpts{
		StyleID: e.CurrentStyle,
	}

	err = e.Writer.SetRow(cell, elements, opts)
	if err != nil {
		return err
	}

	e.CurrentRow += 1

	return nil
}
