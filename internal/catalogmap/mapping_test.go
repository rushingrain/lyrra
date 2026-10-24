package catalogmap

import (
	"testing"

	"lyrra/internal/kgapi"
	"lyrra/internal/play"
)

func TestSearchQuality(t *testing.T) {
	tests := []struct {
		name string
		item kgapi.Record
		want string
	}{
		{
			name: "hi res",
			item: kgapi.Record{"Res": map[string]any{}, "ResDuration": float64(1)},
			want: qualityHiRes,
		},
		{
			name: "lossless",
			item: kgapi.Record{"SQFileHash": "sq-hash", "SQExtName": "flac"},
			want: qualityLossless,
		},
		{
			name: "high quality",
			item: kgapi.Record{"HQFileHash": "hq-hash", "HQBitrate": float64(320)},
			want: qualityHQ,
		},
		{
			name: "standard has no quality badge",
			item: kgapi.Record{"FileHash": "file-hash", "Bitrate": 128},
			want: "",
		},
		{
			name: "missing required fields",
			item: kgapi.Record{"SQBitrate": 702, "HQBitrate": 320},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchQuality(tt.item); got != tt.want {
				t.Fatalf("searchQuality() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLatestSongMapping(t *testing.T) {
	item := kgapi.Record{
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
		t.Fatalf("Hash = %q, want HASH-128", got.Hash)
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
	if got.Quality != qualityHiRes {
		t.Fatalf("Quality = %q, want %q", got.Quality, qualityHiRes)
	}
	if !got.IsOriginal {
		t.Fatal("IsOriginal = false, want true")
	}
}

func TestLatestSongsFiltersFailedRecords(t *testing.T) {
	items := []kgapi.Record{
		{"songname": "失败", "__status": float64(0)},
		{"songname": "成功", "__status": float64(1), "album_audio_id": float64(1)},
	}
	songs := LatestSongs(items, play.PriorityForAny(play.DefaultQuality))
	if len(songs) != 1 || songs[0].Name != "成功" {
		t.Fatalf("LatestSongs() = %+v, want only 成功", songs)
	}
}

func TestRecommendedSongMapping(t *testing.T) {
	item := kgapi.Record{
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
	if got.ID != 32183195 || got.Hash != "HASH-128" {
		t.Fatalf("unexpected identity: %+v", got)
	}
	if got.Name != "冷雨夜 R&B版" {
		t.Fatalf("Name = %q, want 冷雨夜 R&B版", got.Name)
	}
	if got.Artist != "BEYOND" || got.Album != "BTB现代舞台- Beyond" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20250213/cover.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Duration != 259213 || got.Quality != qualityLossless || !got.IsOriginal {
		t.Fatalf("unexpected song fields: %+v", got)
	}
}

func TestDailyRecommendSongMapping(t *testing.T) {
	item := kgapi.Record{
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
			"is_original": float64(2),
		},
	}

	got := parseDailyRecommendSong(item, play.PriorityForAny(play.DefaultQuality))
	if got.ID != 854181271 || got.Name != "情人 (R&B版)" {
		t.Fatalf("unexpected song: %+v", got)
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
	if got.Duration != 237000 || got.Quality != qualityLossless || got.IsOriginal {
		t.Fatalf("unexpected song fields: %+v", got)
	}
}

func TestDailyRecommendSongNameFallback(t *testing.T) {
	item := kgapi.Record{"ori_audio_name": "情人", "mixsongid": "1"}
	if got := parseDailyRecommendSong(item, play.PriorityForAny(play.DefaultQuality)); got.Name != "情人" {
		t.Fatalf("Name = %q, want 情人", got.Name)
	}
}

func TestRecommendedSongHashFallback(t *testing.T) {
	priority := play.PriorityForAny(play.DefaultQuality)

	flacOnly := kgapi.Record{"audio_info": map[string]any{"hash_flac": "HASH-FLAC"}}
	if got := parseRecommendedSong(flacOnly, priority).Hash; got != "HASH-FLAC" {
		t.Fatalf("Hash = %q, want HASH-FLAC", got)
	}

	noAudioInfo := kgapi.Record{"hash": "TOP-LEVEL-HASH"}
	if got := parseRecommendedSong(noAudioInfo, priority).Hash; got != "" {
		t.Fatalf("Hash = %q, want empty", got)
	}

	if got := parseRecommendedSong(kgapi.Record{"ori_audio_name": "冷雨夜"}, priority).Name; got != "冷雨夜" {
		t.Fatalf("Name = %q, want 冷雨夜", got)
	}
}

func TestPlaylistSongMapping(t *testing.T) {
	item := kgapi.Record{
		"name":      "SING女团 - 闹元宵",
		"mixsongid": float64(130794237),
		"timelen":   float64(212032),
		"hash":      "0BD211E917A49ABA51884A12307AB8BF",
		"singerinfo": []any{
			map[string]any{"name": "SING女团"},
		},
		"albuminfo": map[string]any{"name": "闹元宵"},
		"cover":     "http://imge.kugou.com/stdmusic/{size}/20200909/cover.jpg",
		"relate_goods": []any{
			map[string]any{"bitrate": float64(128)},
			map[string]any{"bitrate": float64(320)},
			map[string]any{"bitrate": float64(963)},
		},
		"trans_param": map[string]any{"is_original": float64(1)},
	}

	got := parsePlaylistSong(item, play.PriorityForAny(play.DefaultQuality))
	if got.ID != 130794237 || got.Name != "闹元宵" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.Artist != "SING女团" || got.Album != "闹元宵" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20200909/cover.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Duration != 212032 || got.Hash != "0BD211E917A49ABA51884A12307AB8BF" {
		t.Fatalf("unexpected song fields: %+v", got)
	}
	if got.Quality != qualityLossless || !got.IsOriginal {
		t.Fatalf("unexpected song fields: %+v", got)
	}
}

func TestPlaylistSongsFiltersEmptyNames(t *testing.T) {
	items := []kgapi.Record{
		{"fileid": float64(121), "shield": float64(1)},
		{"name": "SING女团 - 闹元宵", "mixsongid": float64(130794237), "hash": "HASH"},
		{"fileid": float64(120), "shield": float64(1)},
	}
	got := PlaylistSongs(items, play.PriorityForAny(play.DefaultQuality))
	if len(got) != 1 || got[0].Name != "闹元宵" {
		t.Fatalf("PlaylistSongs() = %+v", got)
	}
}

func TestSplitPlaylistSongName(t *testing.T) {
	if artist, name := splitPlaylistSongName("歌手 - 歌曲 - 副标题"); artist != "歌手" || name != "歌曲 - 副标题" {
		t.Fatalf("split = %q / %q", artist, name)
	}
	if artist, name := splitPlaylistSongName("只有歌名"); artist != "" || name != "只有歌名" {
		t.Fatalf("split = %q / %q", artist, name)
	}
}

func TestLyricSongName(t *testing.T) {
	withSuffix := kgapi.Record{
		"SongName":    "一个人想着一个人",
		"trans_param": map[string]any{"songname_suffix": "(说唱版)"},
	}
	if got := lyricSongName(withSuffix); got != "一个人想着一个人 (说唱版)" {
		t.Fatalf("lyricSongName() = %q", got)
	}
	if got := lyricSongName(kgapi.Record{"SongName": "一个人想着一个人"}); got != "一个人想着一个人" {
		t.Fatalf("lyricSongName() = %q", got)
	}
	if got := lyricSongName(kgapi.Record{}); got != "" {
		t.Fatalf("lyricSongName() = %q, want empty", got)
	}
}

func TestLyricQuality(t *testing.T) {
	hq := map[string]any{"Hash": "hq-hash"}
	sq := map[string]any{"Hash": "sq-hash"}
	res := map[string]any{"Hash": "res-hash"}
	tests := []struct {
		name string
		item kgapi.Record
		want string
	}{
		{name: "res", item: kgapi.Record{"HQ": hq, "SQ": sq, "Res": res}, want: qualityHiRes},
		{name: "sq", item: kgapi.Record{"HQ": hq, "SQ": sq}, want: qualityLossless},
		{name: "hq", item: kgapi.Record{"HQ": hq}, want: qualityHQ},
		{name: "empty hash", item: kgapi.Record{"Res": map[string]any{"Hash": ""}, "SQ": map[string]any{}}, want: ""},
		{name: "standard", item: kgapi.Record{"FileHash": "std-hash"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lyricQuality(tt.item); got != tt.want {
				t.Fatalf("lyricQuality() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLatestSongHash(t *testing.T) {
	item := kgapi.Record{
		"songname":    "给未来的自己",
		"author_name": "梁静茹",
		"hash_128":    "standard-hash",
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
			song := parseLatestSong(item, tt.levels)
			if song.Hash != tt.want {
				t.Fatalf("hash = %q, want %q", song.Hash, tt.want)
			}
			if song.Name != "给未来的自己" || song.Artist != "梁静茹" {
				t.Fatalf("unexpected song: %+v", song)
			}
		})
	}
}

func TestLatestSongHashFallback(t *testing.T) {
	priority := play.PriorityForAny(play.DefaultQuality)
	if got := parseLatestSong(kgapi.Record{"hash_flac": "sq-hash"}, priority).Hash; got != "sq-hash" {
		t.Fatalf("hash = %q, want sq-hash", got)
	}
	if got := parseLatestSong(kgapi.Record{"hash": "standard-hash"}, priority).Hash; got != "standard-hash" {
		t.Fatalf("hash = %q, want standard-hash", got)
	}
	if got := parseLatestSong(kgapi.Record{"songname": "无音质"}, priority).Hash; got != "" {
		t.Fatalf("hash = %q, want empty", got)
	}
}
