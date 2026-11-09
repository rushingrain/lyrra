package auth

import (
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestRandomDfid(t *testing.T) {
	first := randomDfid()
	second := randomDfid()

	if len(first) != dfidLength {
		t.Fatalf("length = %d, want %d", len(first), dfidLength)
	}
	for _, char := range first {
		if !strings.ContainsRune(dfidChars, char) {
			t.Fatalf("dfid contains unexpected char %q", char)
		}
	}
	if first == second {
		t.Fatalf("expected two different dfid, got %q twice", first)
	}
}

func TestEnsureDfid(t *testing.T) {
	service, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// 服务端下发的 dfid 原样保留
	service.client.SetCookie("dfid", "server-issued-dfid")
	if got := service.ensureDfid(); got != "server-issued-dfid" {
		t.Fatalf("ensureDfid() = %q, want server-issued-dfid", got)
	}

	// 为空时本地兜底，并写回 cookie 池
	service.client.SetCookie("dfid", "")
	generated := service.ensureDfid()
	if len(generated) != dfidLength {
		t.Fatalf("generated dfid = %q, want %d chars", generated, dfidLength)
	}
	if got := service.client.Cookie()["dfid"]; got != generated {
		t.Fatalf("cookie dfid = %q, want %q", got, generated)
	}

	// SDK 的 "-" 占位同样视作缺失
	service.client.SetCookie("dfid", "-")
	if got := service.ensureDfid(); got == "-" || len(got) != dfidLength {
		t.Fatalf("ensureDfid() = %q, want a generated dfid", got)
	}
}
