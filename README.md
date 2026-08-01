# mapper

[![Go Reference](https://pkg.go.dev/badge/github.com/audunhov/mapper.svg)](https://pkg.go.dev/github.com/audunhov/mapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/audunhov/mapper)](https://goreportcard.com/report/github.com/audunhov/mapper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`mapper` is a flexible, developer-friendly Go library designed to import tabular data (CSV, XLSX) and map/convert it into Go structs (or pointers to structs). It supports:

- 🚀 **Auto-Mapping**: Case-insensitive and tag-based mapping.
- 🛠️ **Explicit Mapping**: Supply a runtime [`ImportMap`](file:///home/audun/code/mapper/mapper.go#L15) to custom map headers to struct fields.
- 📅 **Custom Layouts**: Easily parse dates and times via struct tags (e.g. `layout:2006/01/02`).
- 🔄 **Custom Decoders**: Implements `encoding.TextUnmarshaler` support for parsing custom domain types.
- 🛡️ **Flexible Error Recovery**: Skip invalid rows or run a custom callback per-row using conversion options.

---

## Installation

Add `mapper` to your Go module dependencies:

```bash
go get github.com/audunhov/mapper
```

---

## Quick Start

Here is a basic example of importing a CSV and auto-converting it to Go structs:

```go
package main

import (
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type User struct {
	FirstName string `map:"First Name"` // Explicit tag mapping
	LastName  string                    // Auto-mapped from "last_name"
	Age       int                       // Auto-mapped from "Age"
	IsActive  bool   `map:"active"`     // Auto-mapped from "active"
}

func main() {
	csvData := `First Name,last_name,Age,active
Alice,Smith,30,true
Bob,Jones,25,false`

	// 1. Create CSV importer
	importer := mapper.NewCSVImporter(strings.NewReader(csvData))

	// 2. Parse the CSV source
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	// 3. Convert tabular data directly into Go structs (nil auto-maps fields)
	users, err := mapper.Convert[User](imported, nil)
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		fmt.Printf("%s %s (Age: %d, Active: %t)\n", user.FirstName, user.LastName, user.Age, user.IsActive)
	}
}
```

---

## Advanced Documentation

For detailed walkthroughs and real-world examples:

1. **Detailed Walkthroughs**: [examples.md](file:///home/audun/code/mapper/examples.md)
   - [Basic CSV Import with Auto-Mapping](file:///home/audun/code/mapper/examples.md#1-basic-csv-import-with-auto-mapping)
   - [Explicit Column-to-Field Mapping](file:///home/audun/code/mapper/examples.md#2-explicit-column-to-field-mapping)
   - [Excel (XLSX) Importing](file:///home/audun/code/mapper/examples.md#3-excel-xlsx-importing)
   - [Custom Date & Time Layouts](file:///home/audun/code/mapper/examples.md#4-custom-date--time-layouts)
   - [Custom Type Unmarshalling](file:///home/audun/code/mapper/examples.md#5-custom-type-unmarshalling-textunmarshaler)
   - [Error Handling & Row Skipping Options](file:///home/audun/code/mapper/examples.md#6-error-handling--row-skipping)
2. **Web Importer & Mapping UI**: [web_example.md](file:///home/audun/code/mapper/web_example.md)
   - Creating a visual column mapping UI in web applications.
   - Production tips on memory-efficient file uploads and streaming.

---

## Core API Reference

- [`Convert[T any]`](file:///home/audun/code/mapper/mapper.go#L46): Maps row maps to structural elements of type `T`.
- [`ImportMap`](file:///home/audun/code/mapper/mapper.go#L15): Map of file header strings to Go struct field name strings.
- [`WithSkipOnError`](file:///home/audun/code/mapper/mapper.go#L34): Convenience option to skip rows that fail validation.
- [`WithErrorHandler`](file:///home/audun/code/mapper/mapper.go#L27): Inject callback function to log and skip, or abort parsing per-row.

---

## License

Distributed under the MIT License. See [`LICENSE`](file:///home/audun/code/mapper/LICENSE) for more information.
