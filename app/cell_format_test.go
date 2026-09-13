package app

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// A json or jsonb value arrives from pgx as a Go map or slice, and fmt prints a
// map as "map[k:v ...]". That is what the grid used to show: not JSON, not
// copyable, and not something that survives being written back to the column.
func TestFormatCellValueEncodesJSONColumns(t *testing.T) {
	value := map[string]any{"language": []any{"English", "Spanish"}, "count": 2}

	if got := formatCellValue(value, true); got != `{"count":2,"language":["English","Spanish"]}` {
		t.Errorf("json column: got %q", got)
	}

	// The same value in a column that is not json keeps fmt's rendering, so this
	// change cannot alter how any other type is displayed.
	if got := formatCellValue(value, false); got != "map[count:2 language:[English Spanish]]" {
		t.Errorf("non-json column: got %q", got)
	}
}

func TestFormatCellValueLeavesOtherTypesAlone(t *testing.T) {
	id := uuid.MustParse("0f1e2d3c-4b5a-6978-8796-a5b4c3d2e1f0")
	stamp := time.Date(2026, 9, 13, 22, 17, 48, 0, time.UTC)

	cases := []struct {
		name   string
		value  any
		isJSON bool
		want   string
	}{
		{"nil is the NULL sentinel", nil, true, "NULL"},
		{"empty string is the EMPTY sentinel", "", false, "EMPTY"},
		{"text passes through", "hello", false, "hello"},
		{"bytes become text", []byte(`{"a":1}`), true, `{"a":1}`},
		{"uuid is printed, not marshalled", [16]uint8(id), true, id.String()},
		{"timestamps are RFC3339", stamp, false, "2026-09-13T22:17:48Z"},
		{"numbers keep fmt's rendering", int64(42), false, "42"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatCellValue(tc.value, tc.isJSON); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// A value that cannot be marshalled must still reach the grid rather than
// turning the cell into an empty string.
func TestFormatCellValueFallsBackWhenJSONFails(t *testing.T) {
	unmarshallable := map[string]any{"fn": func() {}}

	if got := formatCellValue(unmarshallable, true); got == "" {
		t.Error("unmarshallable json value produced an empty cell")
	}
}

func TestJSONColumnSetMarksOnlyJSONTypes(t *testing.T) {
	fields := []pgconn.FieldDescription{
		{Name: "id", DataTypeOID: pgtype.Int8OID},
		{Name: "payload", DataTypeOID: pgtype.JSONBOID},
		{Name: "doc", DataTypeOID: pgtype.JSONOID},
		{Name: "tags", DataTypeOID: pgtype.TextArrayOID},
	}

	want := []bool{false, true, true, false}
	got := jsonColumnSet(fields)

	if len(got) != len(want) {
		t.Fatalf("got %d flags, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: got %v, want %v", fields[i].Name, got[i], want[i])
		}
	}
}
