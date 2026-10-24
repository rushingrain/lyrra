package catalogmap

import (
	"testing"

	"lyrra/internal/kgapi"
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
