package search

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	kg "github.com/rushingrain/kugou-music-api"
)

func TestSerializeSearchCookie(t *testing.T) {
	cookie := serializeSearchCookie(map[string]string{
		"token":  "token-value",
		"userid": "123456",
		"dfid":   "dfid-value",
	})

	const want = "token=token-value;userid=123456;dfid=dfid-value"
	if cookie != want {
		t.Fatalf("serializeSearchCookie() = %q, want %q", cookie, want)
	}
}

// TestSearchComplexLive 用当前本地登录会话调用真实接口，便于联调响应结构。
// 该测试需要网络和有效登录态，因此仅在 KUGOU_LIVE_TEST=1 时执行。
func TestSearchComplexLive(t *testing.T) {

	service, err := NewSearchService(slog.New(slog.NewTextHandler(io.Discard, nil)))
	cookie, err := service.loadSearchCookie()
	t.Log(cookie)
	if os.Getenv("KUGOU_LIVE_TEST") != "1" {
		t.Skip("set KUGOU_LIVE_TEST=1 to run the live search test")
	}

	//service, err := NewSearchService(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewSearchService() error = %v", err)
	}

	//cookie, err := service.loadSearchCookie()

	t.Log(cookie)
	if err != nil {
		t.Fatalf("loadSearchCookie() error = %v", err)
	}

	response, err := service.client.SearchComplex(context.Background(), kg.SearchComplexRequest{
		Keywords: "周杰伦",
		Page:     1,
		Pagesize: 6,
		Cookie:   cookie,
		Extra: map[string]any{
			"cookie": serializeSearchCookie(cookie),
		},
	})
	if err != nil {
		t.Fatalf("SDK SearchComplex() error = %v", err)
	}

	t.Logf("SearchComplex http_status=%d headers=%v", response.Status, response.Headers)
	t.Logf("SearchComplex raw_response=%s", response.RawBody)
	t.Logf("SearchComplex status=%v error_code=%v", response.Body["status"], response.Body["error_code"])
	t.Logf("SearchComplex response: %+v", response.Body)
}
