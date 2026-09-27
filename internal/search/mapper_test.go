package search

import "testing"

func TestMapComplexSearch(t *testing.T) {
	result := mapComplexSearch(map[string]any{
		"lists": []any{
			map[string]any{
				"type": "song",
				"lists": []any{
					map[string]any{
						"MixSongID":  "33138036",
						"SongName":   "是时候",
						"SingerName": "孙燕姿",
						"AlbumName":  "史上最伤感的网络情歌",
						"Image":      "https://imge.kugou.com/stdmusic/400/cover.jpg",
						"Duration":   238,
						"SQFileHash": "sq-hash",
						"SQExtName":  "flac",
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
						"specialid":        7215278,
						"specialname":      "是时候",
						"nickname":         "孙燕姿",
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
						"albumid":      2997292,
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

	if len(result.Songs) != 1 || result.Songs[0].ID != 33138036 {
		t.Fatalf("unexpected songs: %+v", result.Songs)
	}
	if result.Songs[0].Duration != 238000 || result.Songs[0].Quality != "无损" {
		t.Fatalf("unexpected song fields: %+v", result.Songs[0])
	}
	if len(result.Playlists) != 1 {
		t.Fatalf("playlist count = %d, want 1", len(result.Playlists))
	}
	playlist := result.Playlists[0]
	if playlist.ID != 7215278 || playlist.Creator != "孙燕姿" || playlist.PlayCount != 1994007 || playlist.SongCount != 40 {
		t.Fatalf("unexpected playlist: %+v", playlist)
	}
	if result.Albums[0].ID != 2997292 || result.Albums[0].Date != "2011-03-08" {
		t.Fatalf("unexpected albums: %+v", result.Albums)
	}
	if len(result.Artists) != 1 || result.Artists[0].ID != 1090689 {
		t.Fatalf("unexpected artists: %+v", result.Artists)
	}
	if len(result.MVs) != 1 || result.MVs[0].ID != 67468 {
		t.Fatalf("unexpected mvs: %+v", result.MVs)
	}
}

func TestMapNormalSearch(t *testing.T) {
	result := mapNormalSearch("song", map[string]any{
		"lists": []any{
			map[string]any{
				"MixSongID":   "33138036",
				"OriSongName": "是时候",
				"SingerName":  "孙燕姿",
				"Image":       "https://example.com/cover.jpg",
				"Duration":    238,
				"HQFileHash":  "hq-hash",
				"FileHash":    "song-hash",
				"HQBitrate":   float64(320),
				"IsOriginal":  1,
				"TagContent":  "热门",
			},
		},
	})

	if len(result.Songs) != 1 {
		t.Fatalf("song count = %d, want 1", len(result.Songs))
	}
	song := result.Songs[0]
	if song.ID != 33138036 || song.Name != "是时候" || song.Artist != "孙燕姿" {
		t.Fatalf("unexpected song: %+v", song)
	}
	if song.Duration != 238000 || song.Quality != "HQ" || !song.IsOriginal {
		t.Fatalf("unexpected song fields: %+v", song)
	}
}

func TestResolveSongQuality(t *testing.T) {
	tests := []struct {
		name string
		item ApiRecord
		want string
	}{
		{
			name: "hi res",
			item: ApiRecord{
				"Res":         map[string]any{},
				"ResDuration": float64(1),
			},
			want: "Hi-Res",
		},
		{
			name: "lossless",
			item: ApiRecord{
				"SQFileHash": "sq-hash",
				"SQExtName":  "flac",
			},
			want: "无损",
		},
		{
			name: "high quality",
			item: ApiRecord{
				"HQFileHash": "hq-hash",
				"HQBitrate":  float64(320),
			},
			want: "HQ",
		},
		{
			name: "standard has no quality badge",
			item: ApiRecord{
				"FileHash": "file-hash",
				"Bitrate":  128,
			},
			want: "",
		},
		{
			name: "missing required fields",
			item: ApiRecord{
				"SQBitrate": 702,
				"HQBitrate": 320,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveSongQuality(tt.item); got != tt.want {
				t.Fatalf("quality = %q, want %q", got, tt.want)
			}
		})
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

func TestMapNormalSearchLyricSongs(t *testing.T) {
	// fixture 取自真实歌词搜索响应（只保留断言用到的字段）。
	result := mapNormalSearch("lyric", map[string]any{
		"lists": []any{
			map[string]any{
				"MixSongID":  float64(666646835),
				"SongName":   "一个人想着一个人",
				"SingerName": "艾比利",
				"TimeLength": float64(189),
				"FileHash":   "A4C9C11B5C6CD698FBA54D0AB3FE71CA",
				"Image":      "http://imge.kugou.com/stdmusic/{size}/20250107/20250107161435476861.jpg",
				"Lyric":      "眼角的泪这不是错觉  有太多话想说 ",
				"trans_param": map[string]any{
					"songname_suffix": "(说唱版)",
				},
				"HQ": map[string]any{
					"Hash":    "8C66FDE0A9658A623EB44218325FC207",
					"BitRate": float64(320),
				},
				"SQ": map[string]any{
					"Hash":    "4E0EF1E5562517F9B89AF360B84DD0FF",
					"BitRate": float64(915),
				},
				"Res": map[string]any{
					"Hash":    "17814477BD57507E7FF6063841C19114",
					"BitRate": float64(1738),
				},
			},
		},
	})

	if len(result.Songs) != 1 {
		t.Fatalf("song count = %d, want 1", len(result.Songs))
	}

	song := result.Songs[0]
	if song.ID != 666646835 || song.Name != "一个人想着一个人 (说唱版)" {
		t.Fatalf("unexpected song: %+v", song)
	}
	if song.Artist != "艾比利" {
		t.Fatalf("artist = %q, want 艾比利", song.Artist)
	}
	if want := "https://imge.kugou.com/stdmusic/400/20250107/20250107161435476861.jpg"; song.CoverURL != want {
		t.Fatalf("cover url = %q, want %q", song.CoverURL, want)
	}
	if song.Duration != 189000 {
		t.Fatalf("duration = %d, want 189000", song.Duration)
	}
	if song.Hash != "A4C9C11B5C6CD698FBA54D0AB3FE71CA" {
		t.Fatalf("hash = %q", song.Hash)
	}
	if song.Quality != "Hi-Res" {
		t.Fatalf("quality = %q, want Hi-Res", song.Quality)
	}
	// textValue 会去掉首尾空白，词之间的连续空格保留。
	if song.MatchedLyric != "眼角的泪这不是错觉  有太多话想说" {
		t.Fatalf("matched lyric = %q", song.MatchedLyric)
	}
}

func TestSongNameByLyric(t *testing.T) {
	tests := []struct {
		name string
		item map[string]any
		want string
	}{
		{
			name: "拼接 trans_param 后缀",
			item: map[string]any{
				"SongName":    "一个人想着一个人",
				"trans_param": map[string]any{"songname_suffix": "(说唱版)"},
			},
			want: "一个人想着一个人 (说唱版)",
		},
		{
			name: "没有 trans_param 时返回原名",
			item: map[string]any{"SongName": "一个人想着一个人"},
			want: "一个人想着一个人",
		},
		{
			name: "后缀为空串时返回原名",
			item: map[string]any{
				"SongName":    "一个人想着一个人",
				"trans_param": map[string]any{"songname_suffix": ""},
			},
			want: "一个人想着一个人",
		},
		{name: "没有歌名时返回空", item: map[string]any{}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := songNameByLyric(tt.item); got != tt.want {
				t.Fatalf("songNameByLyric() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveQualityByNestedHash(t *testing.T) {
	hq := map[string]any{"Hash": "hq-hash"}
	sq := map[string]any{"Hash": "sq-hash"}
	res := map[string]any{"Hash": "res-hash"}

	tests := []struct {
		name string
		item map[string]any
		want string
	}{
		{name: "res 优先", item: map[string]any{"HQ": hq, "SQ": sq, "Res": res}, want: "Hi-Res"},
		{name: "sq 次之", item: map[string]any{"HQ": hq, "SQ": sq}, want: "无损"},
		{name: "只剩 hq", item: map[string]any{"HQ": hq}, want: "HQ"},
		{name: "空 hash 视作不存在", item: map[string]any{"Res": map[string]any{"Hash": ""}, "SQ": map[string]any{}}, want: ""},
		{name: "仅标准音质不展示标签", item: map[string]any{"FileHash": "std-hash"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveQualityByNestedHash(tt.item); got != tt.want {
				t.Fatalf("resolveQualityByNestedHash() = %q, want %q", got, tt.want)
			}
		})
	}
}
