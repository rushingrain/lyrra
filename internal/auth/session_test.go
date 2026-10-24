package auth

import "testing"

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
