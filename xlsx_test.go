package mapper

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestXLSXImportAndConvert(t *testing.T) {
	// Create an in-memory XLSX file
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheetName := "Sheet1"
	// Write headers
	_ = f.SetCellValue(sheetName, "A1", "Name")
	_ = f.SetCellValue(sheetName, "B1", "Age")
	_ = f.SetCellValue(sheetName, "C1", "Active")
	_ = f.SetCellValue(sheetName, "D1", "Percent")

	// Row 1
	_ = f.SetCellValue(sheetName, "A2", "Alice")
	_ = f.SetCellValue(sheetName, "B2", "30")
	_ = f.SetCellValue(sheetName, "C2", "true")
	_ = f.SetCellValue(sheetName, "D2", "98.6")

	// Row 2
	_ = f.SetCellValue(sheetName, "A3", "Bob")
	_ = f.SetCellValue(sheetName, "B3", "25")
	_ = f.SetCellValue(sheetName, "C3", "false")
	_ = f.SetCellValue(sheetName, "D3", "72.3")

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

func TestXLSXSheets(t *testing.T) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	// Sheet1 is created by default. Let's create two more sheets.
	_, _ = f.NewSheet("Metadata")
	_, _ = f.NewSheet("Data")

	// Write data to "Data" sheet
	_ = f.SetCellValue("Data", "A1", "Name")
	_ = f.SetCellValue("Data", "B1", "Age")
	_ = f.SetCellValue("Data", "A2", "Charlie")
	_ = f.SetCellValue("Data", "B2", "40")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("failed to write test excel file: %v", err)
	}

	// Create importer with empty sheet name
	importer := NewXLSXImporter(&buf, "")
	defer func() {
		_ = importer.Close()
	}()

	sheets, err := importer.GetSheets()
	if err != nil {
		t.Fatalf("GetSheets failed: %v", err)
	}

	// Verify we got the list of sheets
	expectedSheets := []string{"Sheet1", "Metadata", "Data"}
	if len(sheets) != len(expectedSheets) {
		t.Fatalf("expected %d sheets, got %d", len(expectedSheets), len(sheets))
	}
	for i, s := range sheets {
		if s != expectedSheets[i] {
			t.Errorf("expected sheet %d to be %q, got %q", i, expectedSheets[i], s)
		}
	}

	// Switch to "Data" sheet
	importer.SetSheet("Data")

	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	if len(imported.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(imported.Rows))
	}

	if imported.Rows[0]["Name"] != "Charlie" || imported.Rows[0]["Age"] != "40" {
		t.Errorf("unexpected imported row: %v", imported.Rows[0])
	}
}
