package song

import (
	"context"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	kg "github.com/rushingrain/kugou-music-api"

	"kuwave/internal/auth"
)

// 内置样例歌曲：《无情画 (愿为你放下万千的荣华)》，来自 TopCard 热门推荐，
// 当前登录态下可以拿到 128 标准音质地址。换歌时用环境变量覆盖，不必改代码。
const defaultPlayURLHash = "63FC9D28DCFC2497C2F0ED8BE8BFAD85"

// TestGetSongPlayURLLive 调用 SDK 的 GetSongPlayURL 并打印解析结果，
// 便于联调不同歌曲、不同音质下的返回。
//
// 该测试需要网络；无损 / Hi-Res 等音质还依赖有效登录态（cookie 取本地会话，
// 未登录时按普通音质请求），因此与其它 live 测试一致，仅在 KUGOU_LIVE_TEST=1 时执行：
//
//	KUGOU_LIVE_TEST=1 go test ./internal/song/ -run TestGetSongPlayURLLive -v
//
// 可选参数（均可用环境变量覆盖，缺省用内置样例）：
//
//	KUGOU_TEST_HASH            歌曲文件 hash（必填项，为空会被 SDK 拒绝）
//	KUGOU_TEST_ALBUM_ID        album_id，缺省 0
//	KUGOU_TEST_ALBUM_AUDIO_ID  album_audio_id，通常是 MixSongID，缺省 0
//	KUGOU_TEST_QUALITY         音质，缺省交给 SDK 兜底为 128
//
// 注意：付费 / 无版权歌曲对当前登录态会返回空地址（GetSongPlayURL 返回 "" 且无 error），
// 这种情况测试会额外打印原始 /song/url 响应用于定位。
func TestGetSongPlayURLLive(t *testing.T) {
	if os.Getenv("KUGOU_LIVE_TEST") != "1" {
		t.Skip("set KUGOU_LIVE_TEST=1 to run the live play url test")
	}

	req := kg.SongPlayURLRequest{
		Hash:         envString("KUGOU_TEST_HASH", defaultPlayURLHash),
		AlbumID:      envInt("KUGOU_TEST_ALBUM_ID", 0),
		AlbumAudioID: envInt("KUGOU_TEST_ALBUM_AUDIO_ID", 0),
		Cookie:       auth.LoadSession(),
		Quality:      "flac",
	}
	if quality := envString("KUGOU_TEST_QUALITY", ""); quality != "" {
		req.Quality = quality
	}

	client, err := kg.New(kg.WithAutoRefresh(false), kg.WithLite(true))
	if err != nil {
		t.Fatalf("kg.New() error = %v", err)
	}

	t.Logf("request: hash=%s albumID=%d albumAudioID=%d quality=%v cookie=%v",
		req.Hash, req.AlbumID, req.AlbumAudioID, req.Quality, cookieKeys(req.Cookie))

	playURL, err := client.GetSongPlayURL(context.Background(), req)
	if err != nil {
		t.Fatalf("GetSongPlayURL() error = %v", err)
	}
	t.Logf("GetSongPlayURL() = %q", playURL)

	// 再取一次明细：命中的是哪个接口（song_url / song_url_new）以及全部候选地址。
	result, err := client.ResolveSongPlayURL(context.Background(), req, kg.WithSongURLAll(true))
	if err != nil {
		t.Fatalf("ResolveSongPlayURL() error = %v", err)
	}
	if result == nil {
		t.Log("ResolveSongPlayURL() = nil：该 hash 没有可用地址（可能是付费歌曲且当前登录态无权限）")
		logRawSongURL(t, client, req)
		return
	}

	t.Logf("source = %s", result.Source)
	for i, url := range result.URLs {
		t.Logf("url[%d] = %s", i, url)
	}
}

// logRawSongURL 打印一次原始 /song/url 响应，用于解释"解析不到地址"的原因
// （付费、无版权、音质不支持等都会表现为 data.url 为空）。
func logRawSongURL(t *testing.T, client *kg.Client, req kg.SongPlayURLRequest) {
	t.Helper()

	raw, err := client.SongUrl(context.Background(), kg.SongUrlRequest{
		Hash:         req.Hash,
		AlbumId:      req.AlbumID,
		AlbumAudioId: req.AlbumAudioID,
		Quality:      req.Quality,
		Cookie:       req.Cookie,
	})
	if err != nil {
		t.Logf("SongUrl() raw error = %v", err)
		return
	}
	if raw == nil {
		t.Log("SongUrl() raw response = nil")
		return
	}
	t.Logf("raw /song/url: status=%v error_code=%v error_msg=%v",
		raw.Body["status"], raw.Body["error_code"], raw.Body["error_msg"])
	t.Logf("raw /song/url data = %v", raw.Body["data"])
}

// cookieKeys 只输出 cookie 的字段名，避免把 token 值写进测试日志。
func cookieKeys(cookie map[string]string) []string {
	keys := make([]string, 0, len(cookie))
	for key, value := range cookie {
		if value != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
