package auth

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLoginCookie(t *testing.T) {
	session := map[string]string{
		"token":     " token ",
		"userid":    " 1001 ",
		"dfid":      "",
		"vip_token": "ignored",
	}

	raw := LoginCookie(session)
	if raw["token"] != " token " || raw["userid"] != " 1001 " {
		t.Fatalf("LoginCookie() = %+v, want original values", raw)
	}
	if _, ok := raw["dfid"]; ok {
		t.Fatalf("LoginCookie() 不应包含空字段: %+v", raw)
	}

	trimmed := TrimmedLoginCookie(session)
	if trimmed["token"] != "token" || trimmed["userid"] != "1001" {
		t.Fatalf("TrimmedLoginCookie() = %+v, want trimmed values", trimmed)
	}
}

func TestNewSessionFileUsesFiveDayExpiry(t *testing.T) {
	now := time.Date(2026, 11, 8, 20, 0, 0, 0, time.Local)
	stored := newSessionFile(map[string]string{
		"token":     "token",
		"userid":    "8.63006722e+08",
		"dfid":      "dfid",
		"vip_token": "vip-token",
		"vip_type":  "1",
	}, now)

	if want := now.Add(5 * 24 * time.Hour).UnixMilli(); stored.ExpiresAt != want {
		t.Fatalf("ExpiresAt = %d, want %d", stored.ExpiresAt, want)
	}
	if stored.UserID != "863006722" {
		t.Fatalf("UserID = %q, want normalized userid", stored.UserID)
	}
}

func TestDecodeSessionLegacyUsesFileTime(t *testing.T) {
	data, err := json.Marshal(map[string]string{"token": "token", "userid": "1001"})
	if err != nil {
		t.Fatal(err)
	}
	fallback := time.Date(2026, 11, 1, 12, 0, 0, 0, time.Local)
	stored, err := decodeSession(data, fallback)
	if err != nil {
		t.Fatal(err)
	}
	if want := fallback.Add(5 * 24 * time.Hour).UnixMilli(); stored.ExpiresAt != want {
		t.Fatalf("ExpiresAt = %d, want %d", stored.ExpiresAt, want)
	}
}

func TestSessionExpired(t *testing.T) {
	expiresAt := time.Date(2026, 11, 8, 20, 0, 0, 0, time.Local).UnixMilli()
	if sessionExpired(expiresAt, time.UnixMilli(expiresAt-1)) {
		t.Fatal("会话在过期前一毫秒不应过期")
	}
	if !sessionExpired(expiresAt, time.UnixMilli(expiresAt)) {
		t.Fatal("会话到达过期时间应立即失效")
	}
}
