package netguard

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

type stubTransport struct {
	calls int
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls++
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Request: req}, nil
}

func newRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%q) error = %v", rawURL, err)
	}
	return req
}

func TestGuardBlocksCleartext(t *testing.T) {
	base := &stubTransport{}
	g := &guard{
		base:    base,
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		allowed: allowedCleartextHosts,
	}

	blocked := []string{
		"http://login.user.kugou.com/v5/login_by_token",
		"http://tracker.kugou.com/v6/priv_url",
		"http://example.com/anything",
	}
	for _, rawURL := range blocked {
		if _, err := g.RoundTrip(newRequest(t, rawURL)); err == nil {
			t.Fatalf("%s 应被拒绝", rawURL)
		}
	}
	if base.calls != 0 {
		t.Fatalf("被拒绝的请求不应到达底层 transport，calls = %d", base.calls)
	}
}

func TestGuardAllowsHTTPSAndLoopback(t *testing.T) {
	base := &stubTransport{}
	g := &guard{
		base:    base,
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		allowed: allowedCleartextHosts,
	}

	allowed := []string{
		"https://gateway.kugou.com/v5/url",
		"https://kugouvip.kugou.com/v1/get_union_vip",
		"http://127.0.0.1:34115/asset.js",
		"http://localhost:8080/session",
		"http://[::1]:9999/media",
	}
	for _, rawURL := range allowed {
		if _, err := g.RoundTrip(newRequest(t, rawURL)); err != nil {
			t.Fatalf("%s 不应被拒绝，got %v", rawURL, err)
		}
	}
	if base.calls != len(allowed) {
		t.Fatalf("底层 transport 调用次数 = %d, want %d", base.calls, len(allowed))
	}
}

func TestGuardAllowlist(t *testing.T) {
	base := &stubTransport{}
	g := &guard{
		base:    base,
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		allowed: map[string]struct{}{"legacy.example.com": {}},
	}

	if _, err := g.RoundTrip(newRequest(t, "http://legacy.example.com/x")); err != nil {
		t.Fatalf("白名单 host 不应被拒绝，got %v", err)
	}
	if _, err := g.RoundTrip(newRequest(t, "http://other.example.com/x")); err == nil {
		t.Fatal("非白名单 host 应被拒绝")
	}
}

func TestIsLoopback(t *testing.T) {
	tests := map[string]bool{
		"localhost":   true,
		"LOCALHOST":   true,
		"localhost.":  true,
		"127.0.0.1":   true,
		"127.1.2.3":   true,
		"::1":         true,
		"example.com": false,
		"10.0.0.1":    false,
		"":            false,
	}
	for host, want := range tests {
		if got := isLoopback(host); got != want {
			t.Fatalf("isLoopback(%q) = %v, want %v", host, got, want)
		}
	}
}
