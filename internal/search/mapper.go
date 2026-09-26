package search

import (
	"fmt"
	"strconv"
	"strings"

	"kuwave/internal/dto"
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
func mapNormalSearch(searchType string, data map[string]any) *Result {
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
func mapComplexSearch(data map[string]any) *Result {
	result := emptySearchResult()

	for _, group := range asRecords(data["lists"]) {
		items := asRecords(group["lists"])
		switch textValue(group["type"]) {
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

func emptySearchResult() Result {
	return Result{
		Songs:     []dto.Song{},
		Playlists: []dto.Playlist{},
		Albums:    []dto.Album{},
		Artists:   []dto.Artist{},
		MVs:       []dto.MV{},
		Lyrics:    []Lyric{},
	}
}

func mapSongs(items []ApiRecord) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		quality := resolveSongQuality(item)
		duration := numberValue(item["Duration"])
		if duration > 0 {
			// 搜索接口返回秒；前端 Song.duration 使用毫秒。
			duration *= 1000
		}

		songs = append(songs, dto.Song{
			ID:         numberValue(item["MixSongID"]),
			Name:       songName(item, "SongName", "OriSongName"),
			Artist:     artistName(item, "SingerName", "Singers"),
			Album:      textValue(item["AlbumName"]),
			CoverURL:   textValue(item["Image"]),
			Duration:   duration,
			Hash:       textValue(item["FileHash"]),
			Quality:    quality,
			IsOriginal: numberValue(item["IsOriginal"]) == 1,
			Tags:       splitTags(textValue(item["TagContent"])),
		})
	}
	return songs
}

func songName(item ApiRecord, nameKey string, oriNameKey string) string {
	if name := textValue(item[nameKey]); name != "" {
		return name
	}

	oriName := textValue(item[oriNameKey])
	suffix := textValue(item["Suffix"])

	if oriName == "" {
		return ""
	}
	if suffix == "" {
		return oriName
	}
	return oriName + " " + suffix
}

// resolveSongQuality 按最高可用音质返回前端展示标签。
// 标准音质不作为品质标签展示，因此命中标准分支时仍返回空。
func resolveSongQuality(item ApiRecord) string {
	// 1. Res 是新版嵌套对象；上游同时会在根级返回 ResDuration。
	_, hasRes := item["Res"].(map[string]interface{})
	if hasRes && numberValue(item["ResDuration"]) > 0 {
		return "Hi-Res"
	}
	// 兼容旧版：Res 字段平铺
	if hash, ok := item["ResFileHash"].(string); ok && hash != "" {
		return "Hi-Res"
	}

	// 2. SQ 无损（新版：嵌套对象）
	if sq, ok := item["SQ"].(map[string]interface{}); ok {
		if hash, ok := sq["Hash"].(string); ok && hash != "" {
			return "无损"
		}
	}
	// 兼容旧版：SQ 字段平铺
	if hash, ok := item["SQFileHash"].(string); ok && hash != "" {
		if ext, ok := item["SQExtName"].(string); ok && ext == "flac" {
			return "无损"
		}
	}

	// 3. HQ（新版：嵌套对象）
	if hq, ok := item["HQ"].(map[string]interface{}); ok {
		if hash, ok := hq["Hash"].(string); ok && hash != "" {
			return "HQ"
		}
	}
	// 兼容旧版：HQ 字段平铺
	if hash, ok := item["HQFileHash"].(string); ok && hash != "" {
		if bitrate, ok := item["HQBitrate"].(float64); ok && bitrate == 320 {
			return "HQ"
		}
	}

	// 4. 标准品质（新旧版一致，都在根层级）
	if hash, ok := item["FileHash"].(string); ok && hash != "" {
		return "" // 标准品质不展示标签
	}

	return ""
}

func mapPlaylists(items []ApiRecord) []dto.Playlist {
	playlists := make([]dto.Playlist, 0, len(items))
	for _, item := range items {
		playlists = append(playlists, dto.Playlist{
			ID:          numberValue(item["specialid"]),
			Name:        textValue(item["specialname"]),
			Creator:     textValue(item["nickname"]),
			CoverURL:    textValue(item["img"]),
			Description: textValue(item["intro"]),
			PlayCount:   numberValue(item["total_play_count"]),
			SongCount:   numberValue(item["song_count"]),
		})
	}
	return playlists
}

func mapAlbums(items []ApiRecord) []dto.Album {
	albums := make([]dto.Album, 0, len(items))
	for _, item := range items {
		albums = append(albums, dto.Album{
			ID:       numberValue(item["albumid"]),
			Name:     textValue(item["albumname"]),
			Artist:   artistName(item, "singer", "singers"),
			CoverURL: textValue(item["img"]),
			Date:     strings.SplitN(textValue(item["publish_time"]), " ", 2)[0],
		})
	}
	return albums
}

func mapArtists(items []ApiRecord) []dto.Artist {
	artists := make([]dto.Artist, 0, len(items))
	for _, item := range items {
		artists = append(artists, dto.Artist{
			ID:        numberValue(item["AuthorId"]),
			Name:      textValue(item["AuthorName"]),
			AvatarURL: textValue(item["Avatar"]),
			SongCount: numberValue(item["AudioCount"]),
			Genre:     textValue(item["Genre"]),
			Fans:      numberValue(item["FansNum"]),
		})
	}
	return artists
}

func mapMVs(items []ApiRecord) []dto.MV {
	mvs := make([]dto.MV, 0, len(items))
	for _, item := range items {
		mvs = append(mvs, dto.MV{
			ID:       numberValue(item["MvID"]),
			Name:     textValue(item["MvName"]),
			Artist:   artistName(item, "SingerName", "Singers"),
			CoverURL: textValue(item["Pic"]),
		})
	}
	return mvs
}

// mapNormalMVs 处理普通搜索 MV 封面的特殊返回格式。
// 该接口的 Pic 可能只返回 "20240118194600124028.jpg" 这类文件名，
// 需要转换为酷狗 MV 图床的完整地址；综合搜索返回的已是完整 URL，不受影响。
func mapNormalMVs(items []ApiRecord) []dto.MV {
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
			ID:     numberValue(item["LyricID"]),
			Title:  textValue(item["SongName"]),
			Artist: artistName(item, "SingerName", "Singers"),
			Album:  "",
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

func artistName(item ApiRecord, nameKey string, singersKey string) string {
	if name := textValue(item[nameKey]); name != "" {
		return name
	}

	singers, _ := item[singersKey].([]any)
	if len(singers) == 0 {
		return ""
	}
	singer, _ := singers[0].(map[string]any)
	return textValue(singer["name"])
}

func textValue(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func numberValue(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int:
		return int64(typed)
	case int64:
		return typed
	case string:
		number, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0
		}
		return number
	default:
		return 0
	}
}
