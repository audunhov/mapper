package mapper

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

type XLSXImporter struct {
	reader   io.Reader
	sheet    string
	imported *Imported
	err      error
}

func NewXLSXImporter(reader io.Reader, sheet string) *XLSXImporter {
	return &XLSXImporter{
		reader: reader,
		sheet:  sheet,
	}
}

func (xi *XLSXImporter) parse() error {
	if xi.imported != nil || xi.err != nil {
		return xi.err
	}

	f, err := excelize.OpenReader(xi.reader)
	if err != nil {
		xi.err = fmt.Errorf("failed to parse xlsx: %w", err)
		return xi.err
	}
	defer f.Close()

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

	rows, err := f.GetRows(sheetName)
	if err != nil {
		xi.err = fmt.Errorf("failed to get rows for sheet %q: %w", sheetName, err)
		return xi.err
	}

	if len(rows) == 0 {
		xi.imported = &Imported{}
		return nil
	}

	headers := rows[0]
	var importedRows []map[string]string

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowmap := make(map[string]string)
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

