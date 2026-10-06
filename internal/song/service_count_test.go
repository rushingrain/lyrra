package song

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	kg "github.com/rushingrain/kugou-music-api"
)

// stubClient 记录调用参数并可返回预设响应，避免单测打真实网络。
type stubClient struct {
	favoriteCalls int
	favoriteReq   kg.FavoriteCountRequest
	favoriteResp  *kg.FavoriteCountResponse
	favoriteErr   error

	commentCalls int
	commentReq   kg.CommentCountRequest
	commentResp  *kg.CommentCountResponse
	commentErr   error
}

func (s *stubClient) TopSong(context.Context, kg.TopSongRequest) (*kg.TopSongResponse, error) {
	return nil, errors.New("TopSong 不应被计数接口调用")
}

func (s *stubClient) TopCard(context.Context, kg.TopCardRequest) (*kg.TopCardResponse, error) {
	return nil, errors.New("TopCard 不应被计数接口调用")
}

func (s *stubClient) FavoriteCount(_ context.Context, req kg.FavoriteCountRequest) (*kg.FavoriteCountResponse, error) {
	s.favoriteCalls++
	s.favoriteReq = req
	return s.favoriteResp, s.favoriteErr
}

func (s *stubClient) CommentCount(_ context.Context, req kg.CommentCountRequest) (*kg.CommentCountResponse, error) {
	s.commentCalls++
	s.commentReq = req
	return s.commentResp, s.commentErr
}

// newCountService 构造只注入桩 client 的被测服务。
func newCountService(client songClient) *Service {
	return &Service{client: client, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

// favoriteBody 构造收藏数响应体，结构取自实测：
// {"errcode":0,"errmsg":"","data":{"list":[...]},"status":1}
func favoriteBody(entries ...map[string]any) map[string]any {
	list := make([]any, 0, len(entries))
	for _, entry := range entries {
		list = append(list, entry)
	}
	return map[string]any{
		"errcode": float64(0),
		"errmsg":  "",
		"data":    map[string]any{"list": list},
		"status":  float64(1),
	}
}

func TestGetFavoriteCount(t *testing.T) {
	client := &stubClient{favoriteResp: &kg.FavoriteCountResponse{Body: favoriteBody(map[string]any{
		"mixsongid":  float64(27770044),
		"count":      float64(8552140),
		"count_text": "855w",
	})}}
	service := newCountService(client)

	count, err := service.getFavoriteCount(27770044)
	if err != nil {
		t.Fatalf("getFavoriteCount() error = %v", err)
	}
	if count.Count != 8552140 || count.Text != "855w" {
		t.Fatalf("count = %+v, want {8552140 855w}", count)
	}
	if got := fmt.Sprintf("%v", client.favoriteReq.Mixsongids); got != "27770044" {
		t.Fatalf("mixsongids = %q, want 27770044", got)
	}
}

func TestGetFavoriteCountPicksMatchingEntry(t *testing.T) {
	client := &stubClient{favoriteResp: &kg.FavoriteCountResponse{Body: favoriteBody(
		map[string]any{"mixsongid": float64(1), "count": float64(11), "count_text": "11"},
		map[string]any{"mixsongid": float64(2), "count": float64(22), "count_text": "22"},
	)}}
	service := newCountService(client)

	count, err := service.getFavoriteCount(2)
	if err != nil {
		t.Fatalf("getFavoriteCount() error = %v", err)
	}
	if count.Count != 22 || count.Text != "22" {
		t.Fatalf("count = %+v, want {22 22}", count)
	}
}

func TestGetFavoriteCountKeepsSoleEntryWhenIDDiffers(t *testing.T) {
	// 只请求一首歌时上游只回一条；即便 id 字段格式对不上，也按这条处理。
	client := &stubClient{favoriteResp: &kg.FavoriteCountResponse{Body: favoriteBody(map[string]any{
		"mixsongid": "27770044",
		"count":     "8552140",
	})}}
	service := newCountService(client)

	count, err := service.getFavoriteCount(999)
	if err != nil {
		t.Fatalf("getFavoriteCount() error = %v", err)
	}
	if count.Count != 8552140 {
		t.Fatalf("count = %+v, want 8552140", count)
	}
}

func TestGetFavoriteCountEmptyList(t *testing.T) {
	client := &stubClient{favoriteResp: &kg.FavoriteCountResponse{Body: favoriteBody()}}
	service := newCountService(client)

	count, err := service.getFavoriteCount(27770044)
	if err != nil {
		t.Fatalf("getFavoriteCount() error = %v", err)
	}
	if count.Count != 0 || count.Text != "0" {
		t.Fatalf("count = %+v, want {0 0}", count)
	}
}

func TestGetFavoriteCountBusinessFailure(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
		want string
	}{
		{
			name: "status 不为 1",
			body: map[string]any{"status": float64(0), "errcode": float64(0)},
			want: "收藏数接口业务失败",
		},
		{
			name: "errcode 非 0",
			body: map[string]any{"status": float64(1), "errcode": float64(20001), "errmsg": "参数错误"},
			want: "20001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newCountService(&stubClient{favoriteResp: &kg.FavoriteCountResponse{Body: tt.body}})
			if _, err := service.getFavoriteCount(27770044); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want contains %q", err, tt.want)
			}
		})
	}
}

func TestGetFavoriteCountSkipsRequestWithoutSongID(t *testing.T) {
	client := &stubClient{}
	service := newCountService(client)

	if _, err := service.getFavoriteCount(0); err == nil {
		t.Fatal("歌曲 ID 无效时应返回错误")
	}
	if client.favoriteCalls != 0 {
		t.Fatalf("参数无效时不应请求上游，实际 %d 次", client.favoriteCalls)
	}
}

func TestGetCommentCount(t *testing.T) {
	const hash = "63FC9D28DCFC2497C2F0ED8BE8BFAD85"
	client := &stubClient{commentResp: &kg.CommentCountResponse{Body: map[string]any{
		hash: float64(80488),
	}}}
	service := newCountService(client)

	count, err := service.getCommentCount(hash)
	if err != nil {
		t.Fatalf("getCommentCount() error = %v", err)
	}
	if count.Count != 80488 || count.Text != "8w" {
		t.Fatalf("count = %+v, want {80488 8w}", count)
	}
	if got := fmt.Sprintf("%v", client.commentReq.Hash); got != hash {
		t.Fatalf("hash = %q, want %q", got, hash)
	}
}

func TestGetCommentCountIgnoresHashCase(t *testing.T) {
	client := &stubClient{commentResp: &kg.CommentCountResponse{Body: map[string]any{
		"63FC9D28DCFC2497C2F0ED8BE8BFAD85": float64(80488),
	}}}
	service := newCountService(client)

	count, err := service.getCommentCount("63fc9d28dcfc2497c2f0ed8be8bfad85")
	if err != nil {
		t.Fatalf("getCommentCount() error = %v", err)
	}
	if count.Count != 80488 {
		t.Fatalf("count = %+v, want 80488", count)
	}
}

func TestGetCommentCountMissingHashKey(t *testing.T) {
	// 响应里没有该 hash：按 0 处理，不当成业务失败。
	client := &stubClient{commentResp: &kg.CommentCountResponse{Body: map[string]any{}}}
	service := newCountService(client)

	count, err := service.getCommentCount("63FC9D28DCFC2497C2F0ED8BE8BFAD85")
	if err != nil {
		t.Fatalf("getCommentCount() error = %v", err)
	}
	if count.Count != 0 {
		t.Fatalf("count = %+v, want 0", count)
	}
}

func TestGetCommentCountBusinessFailure(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
		want string
	}{
		{
			name: "status 不为 1",
			body: map[string]any{"status": float64(0), "error": "hash is empty"},
			want: "评论数接口业务失败",
		},
		{
			name: "errcode 非 0",
			body: map[string]any{"errcode": float64(20010), "errmsg": "request params invalid"},
			want: "20010",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newCountService(&stubClient{commentResp: &kg.CommentCountResponse{Body: tt.body}})
			if _, err := service.getCommentCount("63FC9D28DCFC2497C2F0ED8BE8BFAD85"); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want contains %q", err, tt.want)
			}
		})
	}
}

func TestGetCommentCountSkipsRequestWithoutHash(t *testing.T) {
	client := &stubClient{}
	service := newCountService(client)

	if _, err := service.getCommentCount("   "); err == nil {
		t.Fatal("hash 为空时应返回错误")
	}
	if client.commentCalls != 0 {
		t.Fatalf("参数无效时不应请求上游，实际 %d 次", client.commentCalls)
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{in: 0, want: "0"},
		{in: 999, want: "999"},
		{in: 1000, want: "1k"},
		{in: 9999, want: "9k"},
		{in: 10000, want: "1w"},
		{in: 86234, want: "8w"},
		{in: 8552140, want: "855w"},
		{in: 99999999, want: "9999w"},
		{in: 100000000, want: "1亿"},
		{in: 123456789, want: "1亿"},
		{in: 1500000000, want: "15亿"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatCount(tt.in); got != tt.want {
				t.Fatalf("formatCount(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetSongStats(t *testing.T) {
	const hash = "63FC9D28DCFC2497C2F0ED8BE8BFAD85"
	client := &stubClient{
		favoriteResp: &kg.FavoriteCountResponse{Body: favoriteBody(map[string]any{
			"mixsongid": float64(27770044),
			"count":     float64(8552140),
		})},
		commentResp: &kg.CommentCountResponse{Body: map[string]any{hash: float64(80488)}},
	}
	service := newCountService(client)

	stats, err := service.GetSongStats(27770044, hash)
	if err != nil {
		t.Fatalf("GetSongStats() error = %v", err)
	}
	if stats.Favorite == nil || stats.Favorite.Count != 8552140 || stats.Favorite.Text != "855w" {
		t.Fatalf("favorite = %+v, want {8552140 855w}", stats.Favorite)
	}
	if stats.Comment == nil || stats.Comment.Count != 80488 || stats.Comment.Text != "8w" {
		t.Fatalf("comment = %+v, want {80488 8w}", stats.Comment)
	}
}

func TestGetSongStatsKeepsPartialResult(t *testing.T) {
	const hash = "63FC9D28DCFC2497C2F0ED8BE8BFAD85"
	client := &stubClient{
		favoriteResp: &kg.FavoriteCountResponse{Body: favoriteBody(map[string]any{
			"mixsongid": float64(27770044),
			"count":     float64(8552140),
		})},
		// 评论数业务失败：只影响自己那一项
		commentResp: &kg.CommentCountResponse{Body: map[string]any{"status": float64(0)}},
	}
	service := newCountService(client)

	stats, err := service.GetSongStats(27770044, hash)
	if err != nil {
		t.Fatalf("单项失败不应整体报错: %v", err)
	}
	if stats.Favorite == nil || stats.Favorite.Text != "855w" {
		t.Fatalf("favorite = %+v, want 855w", stats.Favorite)
	}
	if stats.Comment != nil {
		t.Fatalf("评论数取不到时应置空，实际 %+v", stats.Comment)
	}
}

func TestGetSongStatsFailsWhenBothUnavailable(t *testing.T) {
	client := &stubClient{
		favoriteErr: errors.New("boom"),
		commentErr:  errors.New("boom"),
	}
	service := newCountService(client)

	if _, err := service.GetSongStats(27770044, "63FC9D28DCFC2497C2F0ED8BE8BFAD85"); err == nil {
		t.Fatal("两项都取不到时应返回错误")
	}
}

func TestGetSongStatsSkipsRequestWithoutInputs(t *testing.T) {
	client := &stubClient{}
	service := newCountService(client)

	if _, err := service.GetSongStats(0, "  "); err == nil {
		t.Fatal("入参都不合法时应返回错误")
	}
	if client.favoriteCalls != 0 || client.commentCalls != 0 {
		t.Fatalf("参数无效时不应请求上游，favorite=%d comment=%d", client.favoriteCalls, client.commentCalls)
	}
}
