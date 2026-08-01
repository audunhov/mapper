package mapper

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestXLSXImportAndConvert(t *testing.T) {
	// Create an in-memory XLSX file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	// Write headers
	f.SetCellValue(sheetName, "A1", "Name")
	f.SetCellValue(sheetName, "B1", "Age")
	f.SetCellValue(sheetName, "C1", "Active")
	f.SetCellValue(sheetName, "D1", "Percent")

	// Row 1
	f.SetCellValue(sheetName, "A2", "Alice")
	f.SetCellValue(sheetName, "B2", "30")
	f.SetCellValue(sheetName, "C2", "true")
	f.SetCellValue(sheetName, "D2", "98.6")

	// Row 2
	f.SetCellValue(sheetName, "A3", "Bob")
	f.SetCellValue(sheetName, "B3", "25")
	f.SetCellValue(sheetName, "C3", "false")
	f.SetCellValue(sheetName, "D3", "72.3")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("failed to write test excel file: %v", err)
	}

	importer := NewXLSXImporter(&buf, "")
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import XLSX: %v", err)
	}

	if len(imported.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(imported.Rows))
	}

	mapping := ImportMap{
		"Name":    "Name",
		"Age":     "Age",
		"Active":  "Active",
		"Percent": "Percent",
	}

	users, err := Convert[TestUser](imported, mapping)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].Name != "Alice" || users[0].Age != 30 || !users[0].Active || users[0].Percent != 98.6 {
		t.Errorf("unexpected Alice: %+v", users[0])
	}

	if users[1].Name != "Bob" || users[1].Age != 25 || users[1].Active || users[1].Percent != 72.3 {
		t.Errorf("unexpected Bob: %+v", users[1])
	}
}
