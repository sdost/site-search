package main

import (
	"encoding/csv"
	"strings"
	"testing"
)

// Row test row struct
type Row struct {
	Str string
	Num int
}

// Row test row struct
type BadRow struct {
	Str  string
	Num  int
	Good bool
}

// TestUnmarshalCSVToReturnData Test unmarshal CSV function
func TestUnmarshalCSVToReturnData(t *testing.T) {
	csvStr := `"one",1
	"two",2
	`
	r := csv.NewReader(strings.NewReader(csvStr))
	var testRow Row
	err := UnmarshalCSV(r, &testRow)

	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if testRow.Str != "one" {
		t.Fatalf("Expected test_row.str to equal 'one', found %s", testRow.Str)
	}

	if testRow.Num != 1 {
		t.Fatalf("Expected test_row.str to equal 1, found %d", testRow.Num)
	}
}

// TestUnmarshalCSVToReturnFieldMismatch Test unmarshal error
func TestUnmarshalCSVToReturnFieldMismatch(t *testing.T) {
	csvStr := `"one"
	"two"
	`
	r := csv.NewReader(strings.NewReader(csvStr))
	var testRow Row
	err := UnmarshalCSV(r, &testRow)

	_, ok := err.(*FieldMismatch)

	if ok == false {
		t.Fatalf("Expected FieldMismatch error, but got %v", err)
	}
}

// TestUnmarshalCSVToReturnUnsupportedType Test unmarshal error
func TestUnmarshalCSVToReturnUnsupportedType(t *testing.T) {
	csvStr := `"one",1,true
	"two",2,false
	`
	r := csv.NewReader(strings.NewReader(csvStr))
	var testRow BadRow
	err := UnmarshalCSV(r, &testRow)

	_, ok := err.(*UnsupportedType)

	if ok == false {
		t.Fatalf("Expected UnsupportedType error, but got %v", err)
	}
}
