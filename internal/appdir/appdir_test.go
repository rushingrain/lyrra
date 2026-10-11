package appdir

import (
	"path/filepath"
	"testing"
)

func TestDirEndsWithAppName(t *testing.T) {
	if got := filepath.Base(Dir()); got != appName {
		t.Fatalf("Dir() 末级目录 = %q, want %q", got, appName)
	}
}
