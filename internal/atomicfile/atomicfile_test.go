package atomicfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCreatesParentAndReplacesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	if err := Write(path, []byte("first")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := Write(path, []byte("second")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("content = %q, want second", data)
	}
}
