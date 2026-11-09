package catalogmap

import (
	"testing"

	"lyrra/internal/kgapi"
	"lyrra/internal/play"
)

func TestHotPlaylistsMapping(t *testing.T) {
	items := []kgapi.Record{{
		"specialid":            float64(1),
		"global_collection_id": "collection-1",
		"specialname":          "歌单",
		"nickname":             "创建者",
		"user_avatar":          "",
		"pic":                  "http://imge.kugou.com/user/{size}/avatar.jpg",
		"tags":                 []any{map[string]any{"tag_name": "流行"}},
		"publishtime":          "2026-01-01",
		"imgurl":               "http://imge.kugou.com/stdmusic/{size}/cover.jpg",
		"intro":                "简介",
		"play_count":           "12",
		"song_count":           float64(3),
	}}
	got := HotPlaylists(items)
	if len(got) != 1 {
		t.Fatalf("HotPlaylists() = %+v", got)
	}
	playlist := got[0]
	if playlist.ID != 1 || playlist.GlobalCollectionID != "collection-1" || playlist.PlayCount != 12 || playlist.SongCount != 3 {
		t.Fatalf("unexpected playlist: %+v", playlist)
	}
	if playlist.CreatorAvatar != "https://imge.kugou.com/user/400/avatar.jpg" {
		t.Fatalf("CreatorAvatar = %q", playlist.CreatorAvatar)
	}
	if playlist.CoverURL != "https://imge.kugou.com/stdmusic/400/cover.jpg" {
		t.Fatalf("CoverURL = %q", playlist.CoverURL)
	}
	if len(playlist.Tags) != 1 || playlist.Tags[0] != "流行" {
		t.Fatalf("Tags = %+v", playlist.Tags)
	}
}

func TestPlaylistInfoMapping(t *testing.T) {
	data := kgapi.Record{
		"list_info": map[string]any{
			"specialid":            float64(2),
			"global_collection_id": "collection-2",
			"name":                 "歌单二",
			"pic":                  "http://imge.kugou.com/stdmusic/{size}/cover.jpg",
			"intro":                "简介",
			"list_create_username": "创建者",
			"create_user_pic":      "http://imge.kugou.com/user/{size}/avatar.jpg",
			"musiclib_tags":        []any{map[string]any{"tag_name": "摇滚"}},
			"publish_date":         "2026-01-02",
			"count":                "4",
		},
	}
	got := PlaylistInfo(data)
	if got.ID != 2 || got.SongCount != 4 || got.Name != "歌单二" {
		t.Fatalf("unexpected playlist: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/cover.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.CreatorAvatar != "https://imge.kugou.com/user/400/avatar.jpg" {
		t.Fatalf("CreatorAvatar = %q", got.CreatorAvatar)
	}
}

func TestAlbumsMappingMultipleArtists(t *testing.T) {
	items := []kgapi.Record{{
		"albumid":      float64(1),
		"albumname":    "合唱专辑",
		"singers":      []any{map[string]any{"name": "歌手A"}, map[string]any{"name": "歌手B"}},
		"img":          "http://imge.kugou.com/stdmusic/{size}/cover.jpg",
		"publish_time": "2026-01-02 00:00:00",
	}}

	got := Albums(items)
	if len(got) != 1 || got[0].Artist != "歌手A/歌手B" {
		t.Fatalf("Albums() = %+v", got)
	}
	if got[0].CoverURL != "https://imge.kugou.com/stdmusic/400/cover.jpg" {
		t.Fatalf("CoverURL = %q", got[0].CoverURL)
	}
}

func TestAlbumDetailMapping(t *testing.T) {
	item := kgapi.Record{
		"album_id":      "2997292",
		"album_name":    "是时候",
		"sizable_cover": "http://imge.kugou.com/stdmusic/{size}/20250221180740317959.jpg",
		"publish_date":  "2011-03-08",
		"intro":         "专辑简介",
		"type":          "录音室专辑",
		"language":      "国语",
		"authors": []any{
			map[string]any{"author_name": "孙燕姿"},
			map[string]any{"author_name": "陈蔚甄"},
		},
	}

	got := AlbumDetail(item)
	if got.ID != 2997292 || got.Name != "是时候" || got.Artist != "孙燕姿/陈蔚甄" {
		t.Fatalf("unexpected album: %+v", got)
	}
	if got.CoverURL != "https://imge.kugou.com/stdmusic/400/20250221180740317959.jpg" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Date != "2011-03-08" || got.Description != "专辑简介" || got.Type != "录音室专辑" || got.Language != "国语" {
		t.Fatalf("unexpected album fields: %+v", got)
	}
}

func TestAlbumSongsMapping(t *testing.T) {
	items := []kgapi.Record{
		{
			"base": map[string]any{
				"album_audio_id": "68434245",
				"audio_name":     "世说心语",
				"author_name":    "孙燕姿",
			},
			"audio_info": map[string]any{
				"hash":      "HASH-128",
				"hash_128":  "HASH-128",
				"hash_320":  "HASH-320",
				"hash_flac": "HASH-FLAC",
				"duration":  222145,
			},
			"album_info": map[string]any{
				"album_name": "是时候",
				"cover":      "http://imge.kugou.com/stdmusic/{size}/20250221180740317959.jpg",
			},
			"authors": []any{
				map[string]any{"author_name": "孙燕姿"},
				map[string]any{"author_name": "陈蔚甄"},
			},
			"trans_param": map[string]any{"is_original": float64(1)},
		},
		{"base": map[string]any{"audio_name": ""}},
	}

	got := AlbumSongs(items, play.PriorityForAny(play.DefaultQuality))
	if len(got) != 1 {
		t.Fatalf("AlbumSongs() = %+v", got)
	}
	song := got[0]
	if song.ID != 68434245 || song.Name != "世说心语" || song.Artist != "孙燕姿/陈蔚甄" {
		t.Fatalf("unexpected song: %+v", song)
	}
	if song.Album != "是时候" || song.Duration != 222145 || song.Hash != "HASH-128" || song.Quality != qualityLossless || !song.IsOriginal {
		t.Fatalf("unexpected song fields: %+v", song)
	}
	if song.CoverURL != "https://imge.kugou.com/stdmusic/400/20250221180740317959.jpg" {
		t.Fatalf("CoverURL = %q", song.CoverURL)
	}
}
