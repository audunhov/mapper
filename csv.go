package mapper

import (
	"encoding/csv"
	"io"
)

type CSVImporter struct {
	reader   io.Reader
	imported *Imported
	err      error
}

func NewCSVImporter(reader io.Reader) *CSVImporter {
	return &CSVImporter{
		reader: reader,
	}
}

func (ci *CSVImporter) parse() error {
	if ci.imported != nil || ci.err != nil {
		return ci.err
	}

	reader := csv.NewReader(ci.reader)
	headers, err := reader.Read()
	if err != nil {
		ci.err = err
		return err
	}

	var rows []map[string]string
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			ci.err = err
			return err
		}

		rowmap := make(map[string]string)
		for i, val := range line {
			if i < len(headers) {
				rowmap[headers[i]] = val
			}
		}
		rows = append(rows, rowmap)
	}

	ci.imported = &Imported{
		Headers: headers,
		Rows:    rows,
	}
	return nil
}

func (ci *CSVImporter) Import() (*Imported, error) {
	if err := ci.parse(); err != nil {
		return nil, err
	}
	return ci.imported, nil
}

func (ci *CSVImporter) ReadHeaders() ([]string, error) {
	if err := ci.parse(); err != nil {
		return nil, err
	}
	return ci.imported.Headers, nil
}

