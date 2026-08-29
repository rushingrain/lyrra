package search

import (
	"fmt"
	"strconv"
	"strings"

	"kuwave/internal/imageurl"
)

// ApiRecord 是 SDK 动态 JSON 响应中的对象。
type ApiRecord = map[string]any

func asRecord(value any) ApiRecord {
	if record, ok := value.(map[string]any); ok {
		return record
	}
	return nil
}

func asRecords(value any) []ApiRecord {
	items, ok := value.([]any)
	if !ok {
		return nil
	}

	records := make([]ApiRecord, 0, len(items))
	for _, item := range items {
		if record := asRecord(item); record != nil {
			records = append(records, record)
		}
	}
	return records
}

func normalSearchItems(data map[string]any) []ApiRecord {
	for _, key := range []string{"lists", "info", "list", "data"} {
		if items := asRecords(data[key]); len(items) > 0 {
			return items
		}
	}
	return nil
}

// mapNormalSearch 将普通搜索的动态 SDK 数据转换为稳定 DTO。
func mapNormalSearch(searchType string, data map[string]any) *SearchResult {
	result := emptySearchResult()
	items := normalSearchItems(data)

	switch searchType {
	case "song":
		result.Songs = mapSongs(items)
	case "special":
		result.Playlists = mapPlaylists(items)
	case "album":
		result.Albums = mapAlbums(items)
	case "author":
		result.Artists = mapArtists(items)
	case "mv":
		result.MVs = mapNormalMVs(items)
	case "lyric":
		result.Lyrics = mapLyrics(items)
	}

	return &result
}

// mapComplexSearch 将综合搜索的分组数据转换为稳定 DTO。
// 上游未展示的分组（program、othersearch 等）会按需逐步补充。
func mapComplexSearch(data map[string]any) *SearchResult {
	result := emptySearchResult()

	for _, group := range asRecords(data["lists"]) {
		items := asRecords(group["lists"])
		switch stringValue(group["type"]) {
		case "song":
			result.Songs = append(result.Songs, mapSongs(items)...)
		case "collect":
			result.Playlists = append(result.Playlists, mapPlaylists(items)...)
		case "album":
			result.Albums = append(result.Albums, mapAlbums(items)...)
		case "author":
			result.Artists = append(result.Artists, mapArtists(items)...)
		case "mv":
			result.MVs = append(result.MVs, mapMVs(items)...)
		}
	}

	return &result
}

func emptySearchResult() SearchResult {
	return SearchResult{
		Songs:     []Song{},
		Playlists: []Playlist{},
		Albums:    []Album{},
		Artists:   []Artist{},
		MVs:       []MV{},
		Lyrics:    []Lyric{},
	}
}

func mapSongs(items []ApiRecord) []Song {
	songs := make([]Song, 0, len(items))
	for _, item := range items {
		sqBitrate := firstNumber(item, "SQBitrate", "sq_bitrate")
		hqBitrate := firstNumber(item, "HQBitrate", "hq_bitrate")
		quality := ""
		switch {
		case sqBitrate > 0:
			quality = "SQ"
		case hqBitrate >= 320:
			quality = "HQ"
		}

		duration := firstNumber(item, "Duration", "duration")
		if duration > 0 {
			// 搜索接口返回秒；前端 Song.duration 使用毫秒。
			duration *= 1000
		}

		songs = append(songs, Song{
			ID:         firstString(item, "ID", "MixSongID", "FileHash", "hash", "audio_id", "Audioid"),
			Title:      firstString(item, "SongName", "FileName", "songname", "filename", "name"),
			Artist:     firstArtistName(item, "SingerName", "singername", "singer"),
			Album:      firstString(item, "AlbumName", "albumname", "album"),
			CoverURL:   firstString(item, "Image", "img", "imgurl", "cover"),
			Duration:   int64(duration),
			Hash:       firstString(item, "FileHash", "hash"),
			Quality:    quality,
			IsOriginal: firstNumber(item, "IsOriginal", "is_original") == 1,
			Tags:       splitTags(firstString(item, "TagContent", "tag_content")),
		})
	}
	return songs
}

func mapPlaylists(items []ApiRecord) []Playlist {
	playlists := make([]Playlist, 0, len(items))
	for _, item := range items {
		playlists = append(playlists, Playlist{
			ID:          firstString(item, "specialid", "special_id", "gid", "id"),
			Name:        firstString(item, "specialname", "special_name", "name"),
			CoverURL:    firstString(item, "img", "imgurl", "cover"),
			Description: firstString(item, "intro", "description", "special_name", "specialname"),
			PlayCount:   int64(firstNumber(item, "total_play_count", "play_count")),
			SongCount:   int64(firstNumber(item, "song_count", "songCount")),
		})
	}
	return playlists
}

func mapAlbums(items []ApiRecord) []Album {
	albums := make([]Album, 0, len(items))
	for _, item := range items {
		albums = append(albums, Album{
			ID:       firstString(item, "albumid", "album_id", "id"),
			Name:     firstString(item, "albumname", "album_name", "title", "name"),
			Artist:   firstArtistName(item, "singer", "SingerName", "singername"),
			CoverURL: firstString(item, "img", "imgurl", "cover"),
			Date:     strings.SplitN(firstString(item, "publish_time", "publishtime", "publish_date"), " ", 2)[0],
		})
	}
	return albums
}

func mapArtists(items []ApiRecord) []Artist {
	artists := make([]Artist, 0, len(items))
	for _, item := range items {
		artists = append(artists, Artist{
			ID:        firstString(item, "AuthorId", "UserId", "author_id", "singerid", "id"),
			Name:      firstString(item, "AuthorName", "authorname", "singer", "name"),
			AvatarURL: firstString(item, "Avatar", "img", "imgurl", "pic"),
			SongCount: int64(firstNumber(item, "AudioCount", "audio_count", "songCount", "song_count")),
			Genre:     firstString(item, "genre", "Genre"),
			Fans:      int64(firstNumber(item, "FansNum", "fans", "fans_num")),
		})
	}
	return artists
}

func mapMVs(items []ApiRecord) []MV {
	mvs := make([]MV, 0, len(items))
	for _, item := range items {
		mvs = append(mvs, MV{
			ID:       firstString(item, "MvID", "MvHash", "mv_id", "hash", "id"),
			Name:     firstString(item, "MvName", "mvname", "songname", "filename", "name"),
			Artist:   firstArtistName(item, "SingerName", "singername", "singer"),
			CoverURL: firstString(item, "Pic", "ErectPic", "img", "imgurl", "cover"),
		})
	}
	return mvs
}

// mapNormalMVs 处理普通搜索 MV 封面的特殊返回格式。
// 该接口的 Pic 可能只返回 "20240118194600124028.jpg" 这类文件名，
// 需要转换为酷狗 MV 图床的完整地址；综合搜索返回的已是完整 URL，不受影响。
func mapNormalMVs(items []ApiRecord) []MV {
	mvs := mapMVs(items)
	for i := range mvs {
		mvs[i].CoverURL = normalizeMVPicFilename(mvs[i].CoverURL)
	}
	return mvs
}

func normalizeMVPicFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" ||
		strings.HasPrefix(value, "http://") ||
		strings.HasPrefix(value, "https://") {
		return value
	}

	filename := strings.TrimPrefix(value, "/")
	if len(filename) < 8 {
		return value
	}
	date := filename[:8]
	for _, r := range date {
		if r < '0' || r > '9' {
			return value
		}
	}

	return imageurl.Normalize(
		fmt.Sprintf("http://imge.kugou.com/mvhdpic/400/%s/%s", date, filename),
	)
}

func mapLyrics(items []ApiRecord) []Lyric {
	lyrics := make([]Lyric, 0, len(items))
	for _, item := range items {
		lyrics = append(lyrics, Lyric{
			ID:     firstString(item, "id", "ID", "hash", "FileHash"),
			Title:  firstString(item, "songname", "SongName", "filename", "FileName", "name"),
			Artist: firstArtistName(item, "singername", "SingerName", "singer"),
			Album:  firstString(item, "album_name", "albumname", "album"),
		})
	}
	return lyrics
}

func splitTags(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '|' || r == ','
	})
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			tags = append(tags, part)
		}
	}
	return tags
}

func firstArtistName(item ApiRecord, keys ...string) string {
	if value := firstString(item, keys...); value != "" {
		return value
	}
	for _, key := range []string{"Singers", "singers"} {
		if singer := firstRecord(asRecords(item[key])); singer != nil {
			if name := stringValue(singer["name"]); name != "" {
				return name
			}
		}
	}
	return ""
}

func firstRecord(items []ApiRecord) ApiRecord {
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

func firstString(item ApiRecord, keys ...string) string {
	for _, key := range keys {
		if value := item[key]; value != nil && value != "" {
			return stringValue(value)
		}
	}
	return ""
}

func firstNumber(item ApiRecord, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := item[key]; ok && value != nil && value != "" {
			if number, err := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64); err == nil {
				return number
			}
		}
	}
	return 0
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}
