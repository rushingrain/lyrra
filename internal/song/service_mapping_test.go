package song

import (
	"testing"

	"lyrra/internal/play"
)

// TestParseLatestSongMapping 覆盖 TopSong 的字段映射：
// ID 取 album_audio_id，Hash 优先取低音质的 hash_128。
func TestParseLatestSongMapping(t *testing.T) {
	item := map[string]any{
		"songname":            "光芒",
		"album_name":          "光芒",
		"album_sizable_cover": "http://imge.kugou.com/stdmusic/{size}/20261016/cover.jpg",
		"timelength":          230000,
		"album_audio_id":      975218502,
		"audio_id":            664396990,
		"hash":                "TOP-LEVEL-HASH",
		"hash_128":            "HASH-128",
		"hash_320":            "HASH-320",
		"hash_flac":           "HASH-FLAC",
		"hash_high":           "HASH-HIGH",
		"authors":             []any{map[string]any{"author_name": "黄子弘凡"}},
		"trans_param":         map[string]any{"is_original": float64(1)},
	}

	got := parseLatestSong(item, play.PriorityForAny(play.DefaultQuality))
	if got.ID != 975218502 {
		t.Fatalf("ID = %d, want album_audio_id 975218502", got.ID)
	}
	if got.Hash != "HASH-128" {
		t.Fatalf("Hash = %q, want HASH-128（优先低音质）", got.Hash)
	}
	if got.Name != "光芒" || got.Artist != "黄子弘凡" || got.Album != "光芒" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20261016/cover.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Duration != 230000 {
		t.Fatalf("Duration = %d, want 230000", got.Duration)
	}
	if got.Quality != "Hi-Res" {
		t.Fatalf("Quality = %q, want Hi-Res", got.Quality)
	}
	if !got.IsOriginal {
		t.Fatalf("IsOriginal = false, want true")
	}
}

// TestParseRecommendedSongMapping 覆盖 TopCardYouth 的嵌套结构：
// 音质 hash / 时长 / 原唱 / 歌名后缀在 audio_info，专辑信息在 album_info，
// 顶层 hash 不参与播放。
func TestParseRecommendedSongMapping(t *testing.T) {
	item := map[string]any{
		"ori_audio_name": "冷雨夜",
		"mixsongid":      32183195,
		"audio_id":       "341080",
		"hash":           "TOP-LEVEL-HASH",
		"cover_pic":      "https://youthimgbssdl.kugou.com/cover.jpg",
		"authors":        []any{map[string]any{"author_name": "BEYOND"}},
		"audio_info": map[string]any{
			"hash":       "HASH-128",
			"hash_128":   "HASH-128",
			"hash_320":   "HASH-320",
			"hash_flac":  "HASH-FLAC",
			"timelength": "259213",
			"trans_param": map[string]any{
				"is_original":     float64(1),
				"songname_suffix": "R&B版",
			},
		},
		"album_info": map[string]any{
			"album_name":    "BTB现代舞台- Beyond",
			"sizable_cover": "http://imge.kugou.com/stdmusic/{size}/20250213/cover.jpg",
		},
	}

	got := parseRecommendedSong(item, play.PriorityForAny(play.DefaultQuality))
	if got.ID != 32183195 {
		t.Fatalf("ID = %d, want mixsongid 32183195", got.ID)
	}
	if got.Hash != "HASH-128" {
		t.Fatalf("Hash = %q, want HASH-128（优先低音质）", got.Hash)
	}
	if got.Name != "冷雨夜 R&B版" {
		t.Fatalf("Name = %q, want 冷雨夜 R&B版（歌名 + 空格 + 后缀）", got.Name)
	}
	if got.Artist != "BEYOND" || got.Album != "BTB现代舞台- Beyond" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20250213/cover.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Duration != 259213 {
		t.Fatalf("Duration = %d, want 259213", got.Duration)
	}
	if got.Quality != "无损" {
		t.Fatalf("Quality = %q, want 无损", got.Quality)
	}
	if !got.IsOriginal {
		t.Fatalf("IsOriginal = false, want true")
	}
}

// TestParseDailyRecommendSongMapping 覆盖每日推荐的顶层结构：
// ID 取 mixsongid，歌名直接用 songname（上游已带版本后缀），不做二次拼接。
func TestParseDailyRecommendSongMapping(t *testing.T) {
	item := map[string]any{
		"songname":       "情人 (R&B版)",
		"ori_audio_name": "情人",
		"mixsongid":      "854181271",
		"album_name":     "情人 (R&B版)",
		"sizable_cover":  "http://imge.kugou.com/stdmusic/{size}/20260401/20260401211241951641.jpg",
		"time_length":    float64(237),
		"hash":           "20DC8DE2125BB2839A1DE8C1B6EFA44A",
		"hash_128":       "20DC8DE2125BB2839A1DE8C1B6EFA44A",
		"hash_320":       "62E714C85D8F7B19FF52693292F2F841",
		"hash_flac":      "02A24BBDBD5134EE901DC330DA434E8E",
		"singerinfo":     []any{map[string]any{"name": "励志豪"}},
		"trans_param": map[string]any{
			"is_original":     float64(2),
			"songname_suffix": "(R&B版)",
		},
	}

	got := parseDailyRecommendSong(item, play.PriorityForAny(play.DefaultQuality))
	if got.ID != 854181271 {
		t.Fatalf("ID = %d, want mixsongid 854181271", got.ID)
	}
	if got.Name != "情人 (R&B版)" {
		t.Fatalf("Name = %q, want 直接使用 songname", got.Name)
	}
	if got.Hash != "20DC8DE2125BB2839A1DE8C1B6EFA44A" {
		t.Fatalf("Hash = %q, want hash_128", got.Hash)
	}
	if got.Artist != "励志豪" || got.Album != "情人 (R&B版)" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20260401/20260401211241951641.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Duration != 237000 {
		t.Fatalf("Duration = %d, want 237000（time_length 秒转毫秒）", got.Duration)
	}
	if got.Quality != "无损" {
		t.Fatalf("Quality = %q, want 无损", got.Quality)
	}
	if got.IsOriginal {
		t.Fatalf("IsOriginal = true, want false")
	}
}

// TestParseDailyRecommendSongNameFallback 覆盖 songname 缺失时回退 ori_audio_name。
func TestParseDailyRecommendSongNameFallback(t *testing.T) {
	item := map[string]any{"ori_audio_name": "情人", "mixsongid": "1"}
	if got := parseDailyRecommendSong(item, play.PriorityForAny(play.DefaultQuality)); got.Name != "情人" {
		t.Fatalf("Name = %q, want 情人", got.Name)
	}
}

// TestParseRecommendedSongHashFallback 确认推荐条目缺标准档时向上兜底，
// 且顶层的 hash 不会被当成音质 hash 用于播放。
func TestParseRecommendedSongHashFallback(t *testing.T) {
	priority := play.PriorityForAny(play.DefaultQuality)

	flacOnly := map[string]any{
		"audio_info": map[string]any{"hash_flac": "HASH-FLAC"},
	}
	if got := parseRecommendedSong(flacOnly, priority).Hash; got != "HASH-FLAC" {
		t.Fatalf("Hash = %q, want HASH-FLAC", got)
	}

	noAudioInfo := map[string]any{"hash": "TOP-LEVEL-HASH"}
	if got := parseRecommendedSong(noAudioInfo, priority).Hash; got != "" {
		t.Fatalf("Hash = %q, want empty（顶层 hash 不是音质 hash）", got)
	}

	if got := parseRecommendedSong(map[string]any{"ori_audio_name": "冷雨夜"}, priority).Name; got != "冷雨夜" {
		t.Fatalf("Name = %q, want 冷雨夜（没有后缀时不拼接）", got)
	}
}
