package mapper

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

type XLSXImporter struct {
	reader   io.Reader
	sheet    string
	file     *excelize.File
	imported *Imported
	err      error
}

func NewXLSXImporter(reader io.Reader, sheet string) *XLSXImporter {
	return &XLSXImporter{
		reader: reader,
		sheet:  sheet,
	}
}

func (xi *XLSXImporter) getFile() (*excelize.File, error) {
	if xi.file != nil {
		return xi.file, nil
	}
	if xi.err != nil {
		return nil, xi.err
	}
	f, err := excelize.OpenReader(xi.reader)
	if err != nil {
		xi.err = fmt.Errorf("failed to parse xlsx: %w", err)
		return nil, xi.err
	}
	xi.file = f
	return f, nil
}

// GetSheets returns a list of sheet names in the XLSX file.
func (xi *XLSXImporter) GetSheets() ([]string, error) {
	f, err := xi.getFile()
	if err != nil {
		return nil, err
	}
	return f.GetSheetList(), nil
}

// SetSheet changes the target sheet to be imported and clears any cached import data.
func (xi *XLSXImporter) SetSheet(sheet string) {
	xi.sheet = sheet
	xi.imported = nil
}

// Close closes the underlying excelize File if it was opened.
func (xi *XLSXImporter) Close() error {
	if xi.file != nil {
		err := xi.file.Close()
		xi.file = nil
		return err
	}
	return nil
}

func (xi *XLSXImporter) parse() error {
	if xi.imported != nil || xi.err != nil {
		return xi.err
	}

	f, err := xi.getFile()
	if err != nil {
		return err
	}

	sheetName := xi.sheet
	if sheetName == "" {
		// Use the first sheet if none specified
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			xi.err = fmt.Errorf("xlsx contains no sheets")
			return xi.err
		}
		sheetName = sheets[0]
	}

	rowsIter, err := f.Rows(sheetName)
	if err != nil {
		xi.err = fmt.Errorf("failed to get rows for sheet %q: %w", sheetName, err)
		return xi.err
	}
	defer rowsIter.Close()

	var headers []string
	if rowsIter.Next() {
		headers, err = rowsIter.Columns()
		if err != nil {
			xi.err = fmt.Errorf("failed to read headers for sheet %q: %w", sheetName, err)
			return xi.err
		}
	}

	if len(headers) == 0 {
		xi.imported = &Imported{}
		return nil
	}

	var importedRows []map[string]string

	for rowsIter.Next() {
		row, err := rowsIter.Columns()
		if err != nil {
			xi.err = fmt.Errorf("failed to read row for sheet %q: %w", sheetName, err)
			return xi.err
		}
		rowmap := make(map[string]string, len(headers))
		for j, val := range row {
			if j < len(headers) {
				rowmap[headers[j]] = val
			}
		}
		importedRows = append(importedRows, rowmap)
	}

	xi.imported = &Imported{
		Headers: headers,
		Rows:    importedRows,
	}
	return nil
}

func (xi *XLSXImporter) Import() (*Imported, error) {
	if err := xi.parse(); err != nil {
		return nil, err
	}
	return xi.imported, nil
}

func (xi *XLSXImporter) ReadHeaders() ([]string, error) {
	if err := xi.parse(); err != nil {
		return nil, err
	}
	return xi.imported.Headers, nil
}
