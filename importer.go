package mapper

// Imported holds the headers and rows parsed from a source.
type Imported struct {
	Headers []string
	Rows    []map[string]string
}

// Importer defines the interface for reading different file formats into Imported data.
type Importer interface {
	Import() (*Imported, error)
	ReadHeaders() ([]string, error)
}
