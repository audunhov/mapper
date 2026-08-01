package mapper

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

type BenchUser struct {
	ID        string  `map:"id"`
	Name      string  `map:"name"`
	Age       int     `map:"age"`
	Active    bool    `map:"active"`
	Score     float64 `map:"score"`
	CreatedAt string  `map:"created_at"`
}

func BenchmarkConvert_20k(b *testing.B) {
	// Prepare 20k rows of data
	rows := make([]map[string]string, 20000)
	for i := 0; i < 20000; i++ {
		rows[i] = map[string]string{
			"id":         "usr-" + strconv.Itoa(i),
			"name":       "User " + strconv.Itoa(i),
			"age":        strconv.Itoa(20 + (i % 50)),
			"active":     strconv.FormatBool(i%2 == 0),
			"score":      strconv.FormatFloat(80.5+float64(i%20)/2.0, 'f', 2, 64),
			"created_at": "2026-08-01T19:32:00Z",
		}
	}
	imported := &Imported{
		Headers: []string{"id", "name", "age", "active", "score", "created_at"},
		Rows:    rows,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Convert[BenchUser](imported, nil)
		if err != nil {
			b.Fatalf("convert failed: %v", err)
		}
	}
}

func BenchmarkCSVImport_20k(b *testing.B) {
	// Prepare a CSV string with 20k rows
	var sb strings.Builder
	sb.WriteString("id,name,age,active,score,created_at\n")
	for i := 0; i < 20000; i++ {
		sb.WriteString("usr-" + strconv.Itoa(i) + ",")
		sb.WriteString("User " + strconv.Itoa(i) + ",")
		sb.WriteString(strconv.Itoa(20+(i%50)) + ",")
		sb.WriteString(strconv.FormatBool(i%2 == 0) + ",")
		sb.WriteString(strconv.FormatFloat(80.5+float64(i%20)/2.0, 'f', 2, 64) + ",")
		sb.WriteString("2026-08-01T19:32:00Z\n")
	}
	csvStr := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		importer := NewCSVImporter(strings.NewReader(csvStr))
		_, err := importer.Import()
		if err != nil {
			b.Fatalf("import failed: %v", err)
		}
	}
}

func BenchmarkXLSXImport_20k(b *testing.B) {
	// Create an in-memory XLSX file with 20k rows
	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Sheet1"
	headers := []string{"id", "name", "age", "active", "score", "created_at"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	for i := 0; i < 20000; i++ {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "usr-"+strconv.Itoa(i))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "User "+strconv.Itoa(i))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), strconv.Itoa(20+(i%50)))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), strconv.FormatBool(i%2 == 0))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), strconv.FormatFloat(80.5+float64(i%20)/2.0, 'f', 2, 64))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "2026-08-01T19:32:00Z")
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		b.Fatalf("failed to write test excel file: %v", err)
	}
	xlsxBytes := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		importer := NewXLSXImporter(bytes.NewReader(xlsxBytes), "")
		_, err := importer.Import()
		if err != nil {
			b.Fatalf("import failed: %v", err)
		}
		importer.Close()
	}
}
