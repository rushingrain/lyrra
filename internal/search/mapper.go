package search

import (
	"fmt"
	"strconv"
	"strings"

	"kuwave/internal/dto"
	"kuwave/internal/imageurl"
	"kuwave/internal/play"
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
// hashPriority 是 Song.Hash 的档位候选顺序，取其中第一个非空 hash（见 internal/play）。
func mapNormalSearch(searchType string, data map[string]any, hashPriority []string) *Result {
	result := emptySearchResult()
	items := normalSearchItems(data)

	switch searchType {
	case "song":
		result.Songs = mapSongs(items, hashPriority)
	case "special":
		result.Playlists = mapPlaylists(items)
	case "album":
		result.Albums = mapAlbums(items)
	case "author":
		result.Artists = mapArtists(items)
	case "mv":
		result.MVs = mapNormalMVs(items)
	case "lyric":
		// 歌词搜索接口实际返回的是歌曲列表，因此填充 Songs。
		result.Songs = mapSongsByLyric(items, hashPriority)
	}

	return &result
}

// mapComplexSearch 将综合搜索的分组数据转换为稳定 DTO。
// 上游未展示的分组（program、othersearch 等）会按需逐步补充。
func mapComplexSearch(data map[string]any, hashPriority []string) *Result {
	result := emptySearchResult()

	for _, group := range asRecords(data["lists"]) {
		items := asRecords(group["lists"])
		switch textValue(group["type"]) {
		case "song":
			result.Songs = append(result.Songs, mapSongs(items, hashPriority)...)
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
	}
}

func mapSongs(items []ApiRecord, hashPriority []string) []dto.Song {
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
			Hash:       play.Resolve(item, hashPriority),
			Quality:    quality,
			IsOriginal: numberValue(item["IsOriginal"]) == 1,
			Tags:       splitTags(textValue(item["TagContent"])),
		})
	}
	return songs
}

// joinSongNameSuffix 用空格拼接歌名与后缀（如 "(说唱版)"），后缀为空时返回原名。
// 单曲搜索的后缀在根级 Suffix，歌词搜索的在 trans_param.songname_suffix，
// 取值位置不同但拼接规则一致，因此两者共用本函数。
func joinSongNameSuffix(name string, suffix string) string {
	if name == "" || suffix == "" {
		return name
	}
	return name + " " + suffix
}

// songName 取单曲搜索的歌名：优先 SongName，缺失时回退 OriSongName 拼根级 Suffix。
func songName(item ApiRecord, nameKey string, oriNameKey string) string {
	if name := textValue(item[nameKey]); name != "" {
		return name
	}
	return joinSongNameSuffix(textValue(item[oriNameKey]), textValue(item["Suffix"]))
}

// songNameByLyric 取歌词命中歌曲的歌名：SongName 拼接 trans_param.songname_suffix。
func songNameByLyric(item ApiRecord) string {
	trans := asRecord(item["trans_param"])
	return joinSongNameSuffix(textValue(item["SongName"]), textValue(trans["songname_suffix"]))
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

// mapSongsByLyric 映射歌词命中的歌曲。
//
// 该接口实际返回的是歌曲列表，Lyric 字段只是命中的歌词片段，因此统一映射为
// Song（由调用方填充 Result.Songs），命中片段放入 Song.MatchedLyric。
// 音质与时长的来源也和单曲搜索不同（见 resolveQualityByNestedHash、TimeLength）。
func mapSongsByLyric(items []ApiRecord, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		var duration int64
		// TimeLength 单位为秒；前端 Song.duration 使用毫秒。
		if seconds := numberValue(item["TimeLength"]); seconds > 0 {
			duration = seconds * 1000
		}

		songs = append(songs, dto.Song{
			ID:           numberValue(item["MixSongID"]),
			Name:         songNameByLyric(item),
			Artist:       artistName(item, "SingerName", "Singers"),
			CoverURL:     imageurl.Normalize(textValue(item["Image"])),
			Duration:     duration,
			Hash:         play.Resolve(item, hashPriority),
			Quality:      resolveQualityByNestedHash(item),
			MatchedLyric: textValue(item["Lyric"]),
		})
	}
	return songs
}

// resolveQualityByNestedHash 按最高可用音质返回展示标签。
//
// 歌词命中歌曲的音质只体现在嵌套对象 HQ/SQ/Res 的 Hash 上，没有 ResDuration
// 等辅助字段，因此只判断 Hash 是否存在（Hash 非空即代表该品质存在）；
// 优先级 Res > SQ > HQ，标准音质（根级 FileHash）不展示标签。
// 与 resolveSongQuality 的区别仅在判定依据，业务规则一致。
func resolveQualityByNestedHash(item ApiRecord) string {
	levels := []struct {
		key   string
		label string
	}{
		{"Res", "Hi-Res"},
		{"SQ", "无损"},
		{"HQ", "HQ"},
	}

	for _, level := range levels {
		nested := asRecord(item[level.key])
		if nested != nil && textValue(nested["Hash"]) != "" {
			return level.label
		}
	}
	return ""
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
