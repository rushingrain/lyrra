package playlist

import (
	"testing"

	"lyrra/internal/play"
)

// TestParsePlaylistSong 覆盖歌单歌曲的字段映射：
// name 是 "歌手 - 歌曲名" 拼接格式，展示名取 " - " 之后的部分；ID 取 mixsongid。
func TestParsePlaylistSong(t *testing.T) {
	item := map[string]any{
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
	if got.ID != 130794237 {
		t.Fatalf("ID = %d, want mixsongid 130794237", got.ID)
	}
	if got.Name != "闹元宵" {
		t.Fatalf("Name = %q, want 闹元宵（去掉歌手前缀）", got.Name)
	}
	if got.Artist != "SING女团" || got.Album != "闹元宵" {
		t.Fatalf("unexpected song: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20200909/cover.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Duration != 212032 {
		t.Fatalf("Duration = %d, want 212032", got.Duration)
	}
	if got.Hash != "0BD211E917A49ABA51884A12307AB8BF" {
		t.Fatalf("Hash = %q", got.Hash)
	}
	if got.Quality != "无损" {
		t.Fatalf("Quality = %q, want 无损", got.Quality)
	}
	if !got.IsOriginal {
		t.Fatal("IsOriginal = false, want true")
	}
}

// TestSplitPlaylistSongName 覆盖歌曲名本身带 " - " 以及没有分隔符的情况。
func TestSplitPlaylistSongName(t *testing.T) {
	if artist, name := splitPlaylistSongName("歌手 - 歌曲 - 副标题"); artist != "歌手" || name != "歌曲 - 副标题" {
		t.Fatalf("split = %q / %q, want 歌手 / 歌曲 - 副标题", artist, name)
	}
	if artist, name := splitPlaylistSongName("只有歌名"); artist != "" || name != "只有歌名" {
		t.Fatalf("split = %q / %q, want 空 / 只有歌名", artist, name)
	}
}

// TestAppendValidSongsSkipsShielded 覆盖上游混入的屏蔽占位条目：
// 只有 fileid / shield、没有歌名的条目不应返回给前端。
func TestAppendValidSongsSkipsShielded(t *testing.T) {
	items := []any{
		map[string]any{"fileid": float64(121), "shield": float64(1)},
		map[string]any{"name": "SING女团 - 闹元宵", "mixsongid": float64(130794237), "hash": "HASH"},
		map[string]any{"fileid": float64(120), "shield": float64(1)},
	}

	got := appendValidSongs(nil, items, play.PriorityForAny(play.DefaultQuality))
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1（屏蔽占位条目应被过滤）", len(got))
	}
	if got[0].Name != "闹元宵" {
		t.Fatalf("Name = %q, want 闹元宵", got[0].Name)
	}
}
