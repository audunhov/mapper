package mapper

import (
	"strings"
	"testing"
)

type TestUser struct {
	Name    string
	Age     int
	Active  bool
	Percent float64
}

func TestCSVImportAndConvert(t *testing.T) {
	csvData := `name,age,active,percent
Alice,30,true,98.6
Bob,25,false,72.3`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import CSV: %v", err)
	}

	if len(imported.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(imported.Rows))
	}

	mapping := ImportMap{
		"name":    "Name",
		"age":     "Age",
		"active":  "Active",
		"percent": "Percent",
	}

	// Test non-pointer elements
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

	// Test pointer elements
	userPtrs, err := Convert[*TestUser](imported, mapping)
	if err != nil {
		t.Fatalf("failed to convert to pointer slice: %v", err)
	}

	if len(userPtrs) != 2 {
		t.Fatalf("expected 2 pointers, got %d", len(userPtrs))
	}

	if userPtrs[0].Name != "Alice" || userPtrs[0].Age != 30 || !userPtrs[0].Active || userPtrs[0].Percent != 98.6 {
		t.Errorf("unexpected Alice pointer: %+v", userPtrs[0])
	}
}

func TestConvertErrors(t *testing.T) {
	imported := &Imported{
		Headers: []string{"name", "age"},
		Rows: []map[string]string{
			{"name": "Alice", "age": "not-an-int"},
		},
	}

	mapping := ImportMap{
		"name": "Name",
		"age":  "Age",
	}

	// Parsing error
	_, err := Convert[TestUser](imported, mapping)
	if err == nil {
		t.Error("expected error parsing invalid int, got nil")
	}

	// Invalid field name error
	badMapping := ImportMap{
		"name": "NonExistentField",
	}
	_, err = Convert[TestUser](imported, badMapping)
	if err == nil {
		t.Error("expected error for non-existent field, got nil")
	}

	// Non-struct T error
	_, err = Convert[int](imported, mapping)
	if err == nil {
		t.Error("expected error for non-struct type parameter, got nil")
	}
}
