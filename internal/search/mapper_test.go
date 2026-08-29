package search

import "testing"

func TestMapComplexSearch(t *testing.T) {
	result := mapComplexSearch(map[string]any{
		"lists": []any{
			map[string]any{
				"type": "song",
				"lists": []any{
					map[string]any{
						"ID":         "33138036",
						"SongName":   "是时候",
						"SingerName": "孙燕姿",
						"AlbumName":  "史上最伤感的网络情歌",
						"Image":      "https://imge.kugou.com/stdmusic/400/cover.jpg",
						"Duration":   238,
						"FileHash":   "song-hash",
						"SQBitrate":  702,
						"IsOriginal": 0,
						"TagContent": "热门|经典",
					},
				},
			},
			map[string]any{
				"type": "collect",
				"lists": []any{
					map[string]any{
						"specialid":        "7215278",
						"specialname":      "是时候",
						"img":              "https://example.com/playlist.jpg",
						"intro":            "简介",
						"total_play_count": "1994007",
						"song_count":       40,
					},
				},
			},
			map[string]any{
				"type": "album",
				"lists": []any{
					map[string]any{
						"albumid":      "2997292",
						"albumname":    "是时候",
						"singer":       "孙燕姿",
						"img":          "https://example.com/album.jpg",
						"publish_time": "2011-03-08 00:00:00",
					},
				},
			},
			map[string]any{
				"type": "author",
				"lists": []any{
					map[string]any{
						"AuthorId":   1090689,
						"AuthorName": "孙燕姿",
						"Avatar":     "https://example.com/avatar.jpg",
						"AudioCount": 15,
						"FansNum":    27,
					},
				},
			},
			map[string]any{
				"type": "mv",
				"lists": []any{
					map[string]any{
						"MvID":       67468,
						"MvName":     "是时候",
						"SingerName": "孙燕姿",
						"Pic":        "https://example.com/mv.jpg",
					},
				},
			},
		},
	})

	if len(result.Songs) != 1 || result.Songs[0].ID != "33138036" {
		t.Fatalf("unexpected songs: %+v", result.Songs)
	}
	if result.Songs[0].Duration != 238000 || result.Songs[0].Quality != "SQ" {
		t.Fatalf("unexpected song fields: %+v", result.Songs[0])
	}
	if len(result.Playlists) != 1 || result.Playlists[0].PlayCount != 1994007 || result.Playlists[0].SongCount != 40 {
		t.Fatalf("unexpected playlists: %+v", result.Playlists)
	}
	if len(result.Albums) != 1 || result.Albums[0].Date != "2011-03-08" {
		t.Fatalf("unexpected albums: %+v", result.Albums)
	}
	if len(result.Artists) != 1 || result.Artists[0].ID != "1090689" {
		t.Fatalf("unexpected artists: %+v", result.Artists)
	}
	if len(result.MVs) != 1 || result.MVs[0].ID != "67468" {
		t.Fatalf("unexpected mvs: %+v", result.MVs)
	}
}

func TestMapNormalSearch(t *testing.T) {
	result := mapNormalSearch("song", map[string]any{
		"lists": []any{
			map[string]any{
				"MixSongID":  "33138036",
				"SongName":   "是时候",
				"SingerName": "孙燕姿",
				"Image":      "https://example.com/cover.jpg",
				"Duration":   238,
				"FileHash":   "song-hash",
				"HQBitrate":  320,
				"IsOriginal": 1,
				"TagContent": "热门",
			},
		},
	})

	if len(result.Songs) != 1 {
		t.Fatalf("song count = %d, want 1", len(result.Songs))
	}
	song := result.Songs[0]
	if song.ID != "33138036" || song.Title != "是时候" || song.Artist != "孙燕姿" {
		t.Fatalf("unexpected song: %+v", song)
	}
	if song.Duration != 238000 || song.Quality != "HQ" || !song.IsOriginal {
		t.Fatalf("unexpected song fields: %+v", song)
	}
}

func TestMapNormalSearchMVCover(t *testing.T) {
	result := mapNormalSearch("mv", map[string]any{
		"lists": []any{
			map[string]any{
				"MvID":       67468,
				"MvName":     "是时候",
				"SingerName": "孙燕姿",
				"Pic":        "20240118194600124028.jpg",
			},
		},
	})

	if len(result.MVs) != 1 {
		t.Fatalf("mv count = %d, want 1", len(result.MVs))
	}
	if got, want := result.MVs[0].CoverURL, "https://imge.kugou.com/mvhdpic/400/20240118/20240118194600124028.jpg"; got != want {
		t.Fatalf("MV cover URL = %q, want %q", got, want)
	}
}
