package song

import (
	"testing"

	"kuwave/internal/play"
)

// TestParseSongHash 验证首页榜单（TopSong / TopCard）的 Hash 也按播放品质偏好解析。
func TestParseSongHash(t *testing.T) {
	t.Cleanup(func() { play.SetPreferredQuality(play.DefaultPreferred) })

	item := map[string]any{
		"songname":    "给未来的自己",
		"author_name": "梁静茹",
		"hash":        "standard-hash",
		"hash_320":    "hq-hash",
		"hash_flac":   "sq-hash",
		"hash_high":   "res-hash",
	}

	tests := []struct {
		quality string
		want    string
	}{
		{quality: play.LevelRes, want: "res-hash"},
		{quality: play.LevelSQ, want: "sq-hash"},
		{quality: "", want: "standard-hash"}, // 默认档位是标准
		{quality: play.LevelHQ, want: "hq-hash"},
		{quality: play.LevelStandard, want: "standard-hash"},
	}

	for _, tt := range tests {
		t.Run("preferred="+tt.quality, func(t *testing.T) {
			play.SetPreferredQuality(tt.quality)
			song := parseSong(item, play.Priority())
			if song.Hash != tt.want {
				t.Fatalf("hash = %q, want %q", song.Hash, tt.want)
			}
			if song.Name != "给未来的自己" || song.Artist != "梁静茹" {
				t.Fatalf("unexpected song: %+v", song)
			}
		})
	}
}

// TestParseSongHashFallback 覆盖降级与"完全没有可用 hash"的情况。
func TestParseSongHashFallback(t *testing.T) {
	t.Cleanup(func() { play.SetPreferredQuality(play.DefaultPreferred) })
	play.SetPreferredQuality(play.LevelSQ)
	priority := play.Priority()

	if got := parseSong(map[string]any{"hash_320": "hq-hash", "hash": "standard-hash"}, priority).Hash; got != "hq-hash" {
		t.Fatalf("hash = %q, want hq-hash（无损缺失时降到 HQ）", got)
	}
	if got := parseSong(map[string]any{"hash": "standard-hash"}, priority).Hash; got != "standard-hash" {
		t.Fatalf("hash = %q, want standard-hash", got)
	}
	if got := parseSong(map[string]any{"songname": "无音质"}, priority).Hash; got != "" {
		t.Fatalf("hash = %q, want empty", got)
	}
}
