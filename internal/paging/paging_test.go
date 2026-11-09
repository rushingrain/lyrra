package paging

import "testing"

func TestNormalize(t *testing.T) {
	page, size := Normalize(0, -1, 15)
	if page != 1 || size != 15 {
		t.Fatalf("Normalize() = %d, %d; want 1, 15", page, size)
	}
	page, size = Normalize(3, 30, 15)
	if page != 3 || size != 30 {
		t.Fatalf("Normalize() = %d, %d; want 3, 30", page, size)
	}
	if got := NormalizeSize(0, 20); got != 20 {
		t.Fatalf("NormalizeSize() = %d, want 20", got)
	}
}
