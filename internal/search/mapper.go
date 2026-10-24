package search

import (
	"lyrra/internal/catalogmap"
	"lyrra/internal/dto"
	"lyrra/internal/kgapi"
)

func normalSearchItems(data map[string]any) []kgapi.Record {
	for _, key := range []string{"lists", "info", "list", "data"} {
		if items := kgapi.Records(data[key]); len(items) > 0 {
			return items
		}
	}
	return nil
}

// mapNormalSearch 将普通搜索的动态 SDK 数据转换为稳定 DTO。
// hashPriority 是 Song.Hash 的档位候选顺序，取其中第一个非空 hash（见 internal/play）。
func mapNormalSearch(searchType string, data map[string]any, hashPriority []string) *Result {
	result := emptySearchResult()
	items := normalSearchItems(data)

	switch searchType {
	case "song":
		result.Songs = catalogmap.SearchSongs(items, hashPriority)
	case "special":
		result.Playlists = catalogmap.SearchPlaylists(items)
	case "album":
		result.Albums = catalogmap.Albums(items)
	case "author":
		result.Artists = catalogmap.Artists(items)
	case "mv":
		result.MVs = catalogmap.NormalSearchMVs(items)
	case "lyric":
		// 歌词搜索接口实际返回的是歌曲列表，因此填充 Songs。
		result.Songs = catalogmap.LyricSongs(items, hashPriority)
	}

	return &result
}

// mapComplexSearch 将综合搜索的分组数据转换为稳定 DTO。
// 上游未展示的分组（program、othersearch 等）会按需逐步补充。
func mapComplexSearch(data map[string]any, hashPriority []string) *Result {
	result := emptySearchResult()

	for _, group := range kgapi.Records(data["lists"]) {
		items := kgapi.Records(group["lists"])
		switch kgapi.Text(group["type"]) {
		case "song":
			result.Songs = append(result.Songs, catalogmap.SearchSongs(items, hashPriority)...)
		case "collect":
			result.Playlists = append(result.Playlists, catalogmap.SearchPlaylists(items)...)
		case "album":
			result.Albums = append(result.Albums, catalogmap.Albums(items)...)
		case "author":
			result.Artists = append(result.Artists, catalogmap.Artists(items)...)
		case "mv":
			result.MVs = append(result.MVs, catalogmap.MVs(items)...)
		}
	}

	return &result
}

func emptySearchResult() Result {
	return Result{
		Songs:     []dto.Song{},
		Playlists: []dto.Playlist{},
		Albums:    []dto.Album{},
		Artists:   []dto.Artist{},
		MVs:       []dto.MV{},
	}
}
