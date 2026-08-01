package mapper

import (
	"encoding/csv"
	"io"
)

type CSVOption func(*csv.Reader)

// WithComma sets the CSV field delimiter.
func WithComma(comma rune) CSVOption {
	return func(r *csv.Reader) {
		r.Comma = comma
	}
}

// WithComment sets the comment character.
func WithComment(comment rune) CSVOption {
	return func(r *csv.Reader) {
		r.Comment = comment
	}
}

// WithLazyQuotes configures the reader to allow lazy quotes.
func WithLazyQuotes(lazy bool) CSVOption {
	return func(r *csv.Reader) {
		r.LazyQuotes = lazy
	}
}

// WithTrimLeadingSpace configures the reader to trim leading whitespace.
func WithTrimLeadingSpace(trim bool) CSVOption {
	return func(r *csv.Reader) {
		r.TrimLeadingSpace = trim
	}
}

// WithFieldsPerRecord sets the expected number of fields per record.
func WithFieldsPerRecord(num int) CSVOption {
	return func(r *csv.Reader) {
		r.FieldsPerRecord = num
	}
}

type CSVImporter struct {
	reader   io.Reader
	imported *Imported
	err      error
	options  []CSVOption
}

func NewCSVImporter(reader io.Reader, opts ...CSVOption) *CSVImporter {
	return &CSVImporter{
		reader:  reader,
		options: opts,
	}
}

func (ci *CSVImporter) parse() error {
	if ci.imported != nil || ci.err != nil {
		return ci.err
	}

	reader := csv.NewReader(ci.reader)
	for _, opt := range ci.options {
		opt(reader)
	}

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

