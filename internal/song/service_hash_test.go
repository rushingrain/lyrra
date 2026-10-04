package song

import (
	"testing"

	"kuwave/internal/play"
)

// TestParseSongHash 验证首页榜单（TopSong / TopCard）的 Hash 取值：
// 默认候选顺序优先标准档，指定候选档位时按顺序取第一个非空值。
func TestParseSongHash(t *testing.T) {
	item := map[string]any{
		"songname":    "给未来的自己",
		"author_name": "梁静茹",
		"hash":        "standard-hash",
		"hash_320":    "hq-hash",
		"hash_flac":   "sq-hash",
		"hash_high":   "res-hash",
	}

	tests := []struct {
		name   string
		levels []string
		want   string
	}{
		{name: "默认顺序优先标准档", levels: play.PriorityForAny(play.DefaultQuality), want: "standard-hash"},
		{name: "只给无损档", levels: []string{play.LevelSQ}, want: "sq-hash"},
		{name: "HQ 优先于无损", levels: []string{play.LevelHQ, play.LevelSQ}, want: "hq-hash"},
		{name: "只给 Hi-Res 档", levels: []string{play.LevelRes}, want: "res-hash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			song := parseSong(item, tt.levels)
			if song.Hash != tt.want {
				t.Fatalf("hash = %q, want %q", song.Hash, tt.want)
			}
			if song.Name != "给未来的自己" || song.Artist != "梁静茹" {
				t.Fatalf("unexpected song: %+v", song)
			}
		})
	}
}

// TestParseSongHashFallback 覆盖缺标准档时向上兜底与"完全没有可用 hash"的情况。
// hash 只用于定位歌曲，播放文件由请求里的 quality 参数决定，因此有任意一档就够用。
func TestParseSongHashFallback(t *testing.T) {
	priority := play.PriorityForAny(play.DefaultQuality)

	if got := parseSong(map[string]any{"hash_flac": "sq-hash"}, priority).Hash; got != "sq-hash" {
		t.Fatalf("hash = %q, want sq-hash（缺标准档时向上兜底）", got)
	}
	if got := parseSong(map[string]any{"hash": "standard-hash"}, priority).Hash; got != "standard-hash" {
		t.Fatalf("hash = %q, want standard-hash", got)
	}
	if got := parseSong(map[string]any{"songname": "无音质"}, priority).Hash; got != "" {
		t.Fatalf("hash = %q, want empty", got)
	}
}
