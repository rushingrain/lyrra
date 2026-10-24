package kgapi

import (
	"encoding/json"
	"testing"
)

func TestTextAndScalar(t *testing.T) {
	if got := Text("  歌名  "); got != "歌名" {
		t.Fatalf("Text() = %q, want 歌名", got)
	}
	if got := Text(float64(1)); got != "" {
		t.Fatalf("Text(number) = %q, want 空串", got)
	}

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: ""},
		{name: "trim string", value: " 24 ", want: "24"},
		{name: "nil string", value: "<nil>", want: ""},
		{name: "integral float", value: float64(24), want: "24"},
		{name: "fractional float", value: 1.5, want: "1.5"},
		{name: "map", value: map[string]any{"value": 1}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Scalar(tt.value); got != tt.want {
				t.Fatalf("Scalar() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInt64OK(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		want   int64
		wantOK bool
	}{
		{name: "float", value: float64(12.8), want: 12, wantOK: true},
		{name: "integer string", value: " 42 ", want: 42, wantOK: true},
		{name: "decimal string", value: "8.6e8", want: 860000000, wantOK: true},
		{name: "json number", value: json.Number("17"), want: 17, wantOK: true},
		{name: "invalid", value: "abc", wantOK: false},
		{name: "nil", value: nil, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Int64OK(tt.value)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("Int64OK() = %d, %v; want %d, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestRecordsAndStatus(t *testing.T) {
	value := []any{
		map[string]any{"id": 1},
		"ignored",
		map[string]any{"id": 2},
	}
	records := Records(value)
	if len(records) != 2 || Int(records[1]["id"]) != 2 {
		t.Fatalf("Records() = %+v", records)
	}
	if got := Int(FirstRecord(value)["id"]); got != 1 {
		t.Fatalf("FirstRecord() id = %d, want 1", got)
	}
	if !StatusOK(Record{"status": float64(1)}) {
		t.Fatal("StatusOK() = false, want true")
	}
	if StatusOK(Record{"status": float64(0)}) {
		t.Fatal("StatusOK() = true, want false")
	}
}

func TestScalarInt(t *testing.T) {
	if got := ScalarInt("24"); got != 24 {
		t.Fatalf("ScalarInt() = %d, want 24", got)
	}
	if got := ScalarInt(float64(12)); got != 12 {
		t.Fatalf("ScalarInt() = %d, want 12", got)
	}
	if got := ScalarInt(nil); got != 0 {
		t.Fatalf("ScalarInt(nil) = %d, want 0", got)
	}
}
