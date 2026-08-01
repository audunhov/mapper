package mapper

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

type TestUser struct {
	Name    string
	Age     int
	Active  bool
	Percent float64
}

type CustomStatus string

func (cs *CustomStatus) UnmarshalText(text []byte) error {
	s := string(text)
	if s != "ACTIVE" && s != "INACTIVE" {
		return errors.New("invalid status: " + s)
	}
	*cs = CustomStatus(s)
	return nil
}

type DXUser struct {
	FirstName string       `map:"First Name"`
	LastName  string       // should auto-map from "last_name"
	Age       uint         // testing uint
	Status    CustomStatus `map:"status"`
	CreatedAt time.Time    `map:"created_at"`
	BirthDate time.Time    `map:"birth_date,layout:2006-01-02"`
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

func TestAutoMappingAndDXFeatures(t *testing.T) {
	csvData := `First Name,last_name,Age,status,created_at,birth_date
Alice,Smith,30,ACTIVE,2026-08-01T12:00:00Z,1996-05-15
Bob,Jones,25,INACTIVE,2026-08-01,2001-10-20`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Convert with nil mapping to trigger auto-mapping
	users, err := Convert[DXUser](imported, nil)
	if err != nil {
		t.Fatalf("failed to convert with auto-mapping: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// Verify Alice
	alice := users[0]
	if alice.FirstName != "Alice" {
		t.Errorf("expected FirstName Alice, got %q", alice.FirstName)
	}
	if alice.LastName != "Smith" {
		t.Errorf("expected LastName Smith, got %q", alice.LastName)
	}
	if alice.Age != 30 {
		t.Errorf("expected Age 30, got %d", alice.Age)
	}
	if alice.Status != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %q", alice.Status)
	}
	expectedCreatedAlice, _ := time.Parse(time.RFC3339, "2026-08-01T12:00:00Z")
	if !alice.CreatedAt.Equal(expectedCreatedAlice) {
		t.Errorf("expected CreatedAt %v, got %v", expectedCreatedAlice, alice.CreatedAt)
	}
	expectedBirthAlice, _ := time.Parse("2006-01-02", "1996-05-15")
	if !alice.BirthDate.Equal(expectedBirthAlice) {
		t.Errorf("expected BirthDate %v, got %v", expectedBirthAlice, alice.BirthDate)
	}

	// Verify Bob (with fallback date format for CreatedAt)
	bob := users[1]
	if bob.FirstName != "Bob" {
		t.Errorf("expected FirstName Bob, got %q", bob.FirstName)
	}
	if bob.LastName != "Jones" {
		t.Errorf("expected LastName Jones, got %q", bob.LastName)
	}
	if bob.Age != 25 {
		t.Errorf("expected Age 25, got %d", bob.Age)
	}
	if bob.Status != "INACTIVE" {
		t.Errorf("expected Status INACTIVE, got %q", bob.Status)
	}
	expectedCreatedBob, _ := time.Parse("2006-01-02", "2026-08-01")
	if !bob.CreatedAt.Equal(expectedCreatedBob) {
		t.Errorf("expected CreatedAt %v, got %v", expectedCreatedBob, bob.CreatedAt)
	}
	expectedBirthBob, _ := time.Parse("2006-01-02", "2001-10-20")
	if !bob.BirthDate.Equal(expectedBirthBob) {
		t.Errorf("expected BirthDate %v, got %v", expectedBirthBob, bob.BirthDate)
	}
}

func TestDetailedErrorContext(t *testing.T) {
	csvData := `First Name,last_name,Age,status,created_at,birth_date
Alice,Smith,not-a-number,ACTIVE,2026-08-01T12:00:00Z,1996-05-15`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	_, err = Convert[DXUser](imported, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErrSub := "row 1 (line 2): cannot parse"
	if !strings.Contains(err.Error(), expectedErrSub) {
		t.Errorf("expected error to contain %q, got: %v", expectedErrSub, err)
	}
}

func TestImporterCaching(t *testing.T) {
	csvData := `name,age
Alice,30`
	buf := bytes.NewBufferString(csvData)

	importer := NewCSVImporter(buf)

	// Read headers first
	headers, err := importer.ReadHeaders()
	if err != nil {
		t.Fatalf("ReadHeaders failed: %v", err)
	}
	if len(headers) != 2 || headers[0] != "name" || headers[1] != "age" {
		t.Errorf("unexpected headers: %v", headers)
	}

	// Now import, which should succeed using the cached values
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if len(imported.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(imported.Rows))
	}
	if imported.Rows[0]["name"] != "Alice" || imported.Rows[0]["age"] != "30" {
		t.Errorf("unexpected row data: %v", imported.Rows[0])
	}
}

func TestWithSkipOnError(t *testing.T) {
	csvData := `First Name,last_name,Age,status,created_at,birth_date
Alice,Smith,not-a-number,ACTIVE,2026-08-01T12:00:00Z,1996-05-15
Bob,Jones,25,INACTIVE,2026-08-01,2001-10-20`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Should skip the first row and parse Bob successfully
	users, err := Convert[DXUser](imported, nil, WithSkipOnError())
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].FirstName != "Bob" || users[0].LastName != "Jones" {
		t.Errorf("unexpected user parsed: %+v", users[0])
	}
}

func TestWithErrorHandler_Collect(t *testing.T) {
	csvData := `First Name,last_name,Age,status,created_at,birth_date
Alice,Smith,not-a-number,ACTIVE,2026-08-01T12:00:00Z,1996-05-15
Bob,Jones,25,INACTIVE,2026-08-01,2001-10-20
Charlie,Brown,invalid-age,ACTIVE,2026-08-01,1999-12-12`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	var collectedErrs []error
	users, err := Convert[DXUser](imported, nil, WithErrorHandler(func(err error, rowIndex int, row map[string]string) error {
		collectedErrs = append(collectedErrs, err)
		return nil
	}))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].FirstName != "Bob" {
		t.Errorf("expected Bob, got %s", users[0].FirstName)
	}

	if len(collectedErrs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(collectedErrs))
	}
	if !strings.Contains(collectedErrs[0].Error(), "row 1 (line 2)") {
		t.Errorf("expected first error to be row 1, got: %v", collectedErrs[0])
	}
	if !strings.Contains(collectedErrs[1].Error(), "row 3 (line 4)") {
		t.Errorf("expected second error to be row 3, got: %v", collectedErrs[1])
	}
}

func TestWithErrorHandler_Abort(t *testing.T) {
	csvData := `First Name,last_name,Age,status,created_at,birth_date
Alice,Smith,not-a-number,ACTIVE,2026-08-01T12:00:00Z,1996-05-15
Bob,Jones,25,INACTIVE,2026-08-01,2001-10-20`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	customErr := errors.New("custom abort error")
	_, err = Convert[DXUser](imported, nil, WithErrorHandler(func(err error, rowIndex int, row map[string]string) error {
		return customErr
	}))

	if err != customErr {
		t.Errorf("expected error to be %v, got %v", customErr, err)
	}
}

type PointerUser struct {
	Name    *string       `map:"name"`
	Age     *int          `map:"age"`
	Active  *bool         `map:"active"`
	Percent *float64      `map:"percent"`
	Status  *CustomStatus `map:"status"`
}

func TestPointerFields(t *testing.T) {
	csvData := `name,age,active,percent,status
Alice,30,true,98.6,ACTIVE
Bob,,,,`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import CSV: %v", err)
	}

	users, err := Convert[PointerUser](imported, nil)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// Verify Alice has all fields populated
	alice := users[0]
	if alice.Name == nil || *alice.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %v", alice.Name)
	}
	if alice.Age == nil || *alice.Age != 30 {
		t.Errorf("expected Age 30, got %v", alice.Age)
	}
	if alice.Active == nil || *alice.Active != true {
		t.Errorf("expected Active true, got %v", alice.Active)
	}
	if alice.Percent == nil || *alice.Percent != 98.6 {
		t.Errorf("expected Percent 98.6, got %v", alice.Percent)
	}
	if alice.Status == nil || *alice.Status != "ACTIVE" {
		t.Errorf("expected Status 'ACTIVE', got %v", alice.Status)
	}

	// Verify Bob has nil fields because of empty strings
	bob := users[1]
	if bob.Name == nil || *bob.Name != "Bob" {
		t.Errorf("expected Name 'Bob' even if others are empty, got %v", bob.Name)
	}
	if bob.Age != nil {
		t.Errorf("expected Age to be nil, got %v", *bob.Age)
	}
	if bob.Active != nil {
		t.Errorf("expected Active to be nil, got %v", *bob.Active)
	}
	if bob.Percent != nil {
		t.Errorf("expected Percent to be nil, got %v", *bob.Percent)
	}
	if bob.Status != nil {
		t.Errorf("expected Status to be nil, got %v", *bob.Status)
	}
}

type EmbeddedBase struct {
	ID string `map:"id"`
}

type EmbeddedDetails struct {
	Role string // auto-mapped from "Role"
}

type EmbeddedUser struct {
	EmbeddedBase            // Value-type anonymous struct
	*EmbeddedDetails        // Pointer-type anonymous struct
	Name             string `map:"name"`
}

func TestEmbeddedStructs(t *testing.T) {
	csvData := `id,name,Role
usr-123,Alice,Admin
usr-456,Bob,Editor`

	importer := NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import CSV: %v", err)
	}

	users, err := Convert[EmbeddedUser](imported, nil)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// Verify usr-123 (Alice)
	alice := users[0]
	if alice.ID != "usr-123" {
		t.Errorf("expected ID usr-123, got %s", alice.ID)
	}
	if alice.Name != "Alice" {
		t.Errorf("expected Name Alice, got %s", alice.Name)
	}
	if alice.EmbeddedDetails == nil {
		t.Fatal("expected EmbeddedDetails to be initialized, got nil")
	}
	if alice.Role != "Admin" {
		t.Errorf("expected Role Admin, got %s", alice.Role)
	}

	// Verify usr-456 (Bob)
	bob := users[1]
	if bob.ID != "usr-456" {
		t.Errorf("expected ID usr-456, got %s", bob.ID)
	}
	if bob.Name != "Bob" {
		t.Errorf("expected Name Bob, got %s", bob.Name)
	}
	if bob.EmbeddedDetails == nil {
		t.Fatal("expected EmbeddedDetails to be initialized, got nil")
	}
	if bob.Role != "Editor" {
		t.Errorf("expected Role Editor, got %s", bob.Role)
	}
}

func TestCSVOptions(t *testing.T) {
	// Semicolon delimited file with comment lines
	csvData := `# This is a comment line
name;age
# Another comment
Alice;30
Bob;25`

	importer := NewCSVImporter(
		strings.NewReader(csvData),
		WithComma(';'),
		WithComment('#'),
	)
	imported, err := importer.Import()
	if err != nil {
		t.Fatalf("failed to import CSV with options: %v", err)
	}

	if len(imported.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(imported.Rows))
	}

	users, err := Convert[TestUser](imported, nil)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if users[0].Name != "Alice" || users[0].Age != 30 {
		t.Errorf("unexpected Alice: %+v", users[0])
	}
	if users[1].Name != "Bob" || users[1].Age != 25 {
		t.Errorf("unexpected Bob: %+v", users[1])
	}
}
