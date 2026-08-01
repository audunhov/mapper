# Mapper Examples

`mapper` is a flexible Go library designed to import tabular data (CSV, XLSX) and map/convert it into Go structs (or pointers to structs) with advanced features like automatic column mapping, custom type unmarshalling, custom date parsing layouts, and flexible error recovery options.

This guide provides real-world examples showing how to use the different features of `mapper`.

---

## Table of Contents
1. [Basic CSV Import with Auto-Mapping](#1-basic-csv-import-with-auto-mapping)
2. [Explicit Column-to-Field Mapping](#2-explicit-column-to-field-mapping)
3. [Excel (XLSX) Importing](#3-excel-xlsx-importing)
4. [Custom Date & Time Layouts](#4-custom-date--time-layouts)
5. [Custom Type Unmarshalling (TextUnmarshaler)](#5-custom-type-unmarshalling-textunmarshaler)
6. [Error Handling & Row Skipping](#6-error-handling--row-skipping)
7. [Web-Based Mapping UI (File Upload & Header Mapping)](web_example.md)

---

### 1. Basic CSV Import with Auto-Mapping

By default, if you pass `nil` as the `ImportMap` parameter, `mapper` automatically maps source columns to struct fields by:
1. Looking for explicit struct tags matching the column name (e.g. `map:"First Name"`).
2. Trying exact struct field name matches.
3. Matching normalized names (alphanumeric case-insensitive match, e.g. mapping `"last_name"` or `"LastName"` to the field `LastName`).

```go
package main

import (
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type User struct {
	FirstName string `map:"First Name"` // Explicit tag mapping
	LastName  string                    // Auto-mapped from "last_name" or "LastName"
	Age       int                       // Auto-mapped from "Age"
	IsActive  bool   `map:"active"`     // Auto-mapped from "active"
}

func main() {
	csvData := `First Name,last_name,Age,active
Alice,Smith,30,true
Bob,Jones,25,false`

	// 1. Create the CSV Importer
	importer := mapper.NewCSVImporter(strings.NewReader(csvData))

	// 2. Parse the CSV source into tabular data
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	// 3. Convert tabular data directly into Go structs (pass nil to auto-map)
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

### 2. Explicit Column-to-Field Mapping

If your input source has headers that do not match your struct definition and you do not want to use struct tags, you can provide an explicit `ImportMap`. The `ImportMap` is a runtime dictionary of type `map[string]string` mapping input header keys to exact struct field names.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type Product struct {
	ID    int
	Title string
	Price float64
}

func main() {
	csvData := `sku_id,product_title,unit_price
101,Mechanical Keyboard,99.99
102,Ergonomic Mouse,59.50`

	importer := mapper.NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	// Define explicit mapping from source column header -> struct field name
	mapping := mapper.ImportMap{
		"sku_id":        "ID",
		"product_title": "Title",
		"unit_price":    "Price",
	}

	products, err := mapper.Convert[Product](imported, mapping)
	if err != nil {
		panic(err)
	}

	for _, p := range products {
		fmt.Printf("Product #%d: %s ($%.2f)\n", p.ID, p.Title, p.Price)
	}
}
```

---

### 3. Excel (XLSX) Importing

`mapper` provides `XLSXImporter` to read from Excel files. You can specify a sheet name, or leave it empty to read from the first sheet.

```go
package main

import (
	"fmt"
	"os"

	"github.com/audunhov/mapper"
)

type Employee struct {
	Name       string
	Department string
	Salary     float64
}

func main() {
	file, err := os.Open("employees.xlsx")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create XLSX Importer targeting the first sheet (empty string "")
	// If targeting a specific sheet, pass its name: e.g., "Engineering"
	importer := mapper.NewXLSXImporter(file, "")

	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	employees, err := mapper.Convert[Employee](imported, nil)
	if err != nil {
		panic(err)
	}

	for _, emp := range employees {
		fmt.Printf("%s - %s ($%.2f)\n", emp.Name, emp.Department, emp.Salary)
	}
}
```

---

### 4. Custom Date & Time Layouts

When parsing dates into `time.Time` fields, `mapper` defaults to parsing them as `time.RFC3339` with a fallback to `"2006-01-02"`. You can define custom layout patterns directly in your struct tags using the `layout` option.

```go
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/audunhov/mapper"
)

type Event struct {
	Name string
	// Explicitly specify custom time layouts:
	StartsAt time.Time `map:"starts_at,layout:2006/01/02 15:04"`
	DateOnly time.Time `map:"date_only,layout:02-01-2006"`
}

func main() {
	csvData := `Name,starts_at,date_only
Tech Conference,2026/08/01 09:00,01-08-2026
Webinar,2026/08/15 14:00,15-08-2026`

	importer := mapper.NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	events, err := mapper.Convert[Event](imported, nil)
	if err != nil {
		panic(err)
	}

	for _, e := range events {
		fmt.Printf("Event: %s | Starts: %v | Date: %v\n", e.Name, e.StartsAt, e.DateOnly.Format("2006-01-02"))
	}
}
```

---

### 5. Custom Type Unmarshalling (TextUnmarshaler)

If you have custom domain types (e.g. status enums, IDs, custom structs), you can implement `encoding.TextUnmarshaler` on them. `mapper` will automatically detect the interface and use your implementation during conversion.

```go
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Implement encoding.TextUnmarshaler to validate and customize conversions
func (s *Status) UnmarshalText(text []byte) error {
	val := strings.ToUpper(string(text))
	switch Status(val) {
	case StatusPending, StatusCompleted, StatusFailed:
		*s = Status(val)
		return nil
	default:
		return errors.New("invalid status value: " + val)
	}
}

type Order struct {
	OrderID     string `map:"order_id"`
	OrderStatus Status `map:"status"`
}

func main() {
	csvData := `order_id,status
ORD-001,pending
ORD-002,completed`

	importer := mapper.NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	orders, err := mapper.Convert[Order](imported, nil)
	if err != nil {
		panic(err)
	}

	for _, order := range orders {
		fmt.Printf("Order %s: status is %s\n", order.OrderID, order.OrderStatus)
	}
}
```

---

### 6. Error Handling & Row Skipping

By default, any field parsing or type-conversion error instantly aborts the entire conversion and returns the error along with detailed context (row and line index).

You can control this behavior using conversion options:

#### Option A: Silently Skip Rows with Errors

Use the `WithSkipOnError()` option to ignore row errors and import all other valid rows.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type User struct {
	Name string
	Age  int
}

func main() {
	// Second row has an invalid age, third row is valid
	csvData := `Name,Age
Alice,30
Bob,invalid-age
Charlie,25`

	importer := mapper.NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	// Bob's row will be skipped, Alice and Charlie will be converted
	users, err := mapper.Convert[User](imported, nil, mapper.WithSkipOnError())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully converted %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("- %s (Age: %d)\n", user.Name, user.Age)
	}
}
```

#### Option B: Catch & Collect Errors (Custom Error Handler)

Use `WithErrorHandler(callback)` to log, collect, or conditionally abort conversion. If the callback returns `nil`, the error row is skipped and parsing continues. If it returns a non-nil error, parsing stops immediately.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type User struct {
	Name string
	Age  int
}

func main() {
	csvData := `Name,Age
Alice,30
Bob,invalid-age
Charlie,invalid-age-2
David,28`

	importer := mapper.NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	var collectedErrors []error

	users, err := mapper.Convert[User](imported, nil, mapper.WithErrorHandler(func(err error, rowIndex int, row map[string]string) error {
		// Log or accumulate the error
		collectedErrors = append(collectedErrors, err)
		
		// Return nil to skip this row and continue mapping remaining rows
		return nil
	}))
	
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully converted %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("- %s (Age: %d)\n", user.Name, user.Age)
	}

	fmt.Printf("\nEncountered %d errors during conversion:\n", len(collectedErrors))
	for _, err := range collectedErrors {
		fmt.Printf("  Error: %v\n", err)
	}
}
```

---

### 7. Web-Based Mapping UI (File Upload & Header Mapping)

The complete example of a web-based importer with a mapping user interface, as well as production tips on file caching, streaming uploads, memory optimization, and asynchronous processing, has been moved to its own guide: [web_example.md](web_example.md).


