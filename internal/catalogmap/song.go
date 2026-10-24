// Package catalogmap 将酷狗动态响应映射为跨服务复用的稳定音乐实体。
package catalogmap

import (
	"strings"

	"lyrra/internal/dto"
	"lyrra/internal/imageurl"
	"lyrra/internal/kgapi"
	"lyrra/internal/play"
)

const (
	qualityHiRes    = "Hi-Res"
	qualityLossless = "无损"
	qualityHQ       = "HQ"
)

// SearchSongs 映射普通与综合搜索返回的歌曲。
func SearchSongs(items []kgapi.Record, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		songs = append(songs, dto.Song{
			ID:         kgapi.Int64(item["MixSongID"]),
			Name:       searchSongName(item),
			Artist:     searchArtist(item, "SingerName", "Singers"),
			Album:      kgapi.Text(item["AlbumName"]),
			CoverURL:   kgapi.Text(item["Image"]),
			Duration:   secondsDuration(item["Duration"]),
			Hash:       play.Resolve(item, hashPriority),
			Quality:    searchQuality(item),
			IsOriginal: kgapi.Flag(item["IsOriginal"]),
			Tags:       splitTags(kgapi.Text(item["TagContent"])),
		})
	}
	return songs
}

// LyricSongs 映射歌词搜索结果；该接口实际返回歌曲列表。
func LyricSongs(items []kgapi.Record, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		songs = append(songs, dto.Song{
			ID:           kgapi.Int64(item["MixSongID"]),
			Name:         lyricSongName(item),
			Artist:       searchArtist(item, "SingerName", "Singers"),
			CoverURL:     imageurl.Normalize(kgapi.Text(item["Image"])),
			Duration:     secondsDuration(item["TimeLength"]),
			Hash:         play.Resolve(item, hashPriority),
			Quality:      lyricQuality(item),
			MatchedLyric: kgapi.Text(item["Lyric"]),
		})
	}
	return songs
}

// LatestSongs 映射新歌首发榜，并过滤上游标记失败的条目。
func LatestSongs(items []kgapi.Record, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		if !kgapi.Flag(item["__status"]) {
			continue
		}
		songs = append(songs, parseLatestSong(item, hashPriority))
	}
	return songs
}

// DailyRecommendSongs 映射每日推荐歌曲。
func DailyRecommendSongs(items []kgapi.Record, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		songs = append(songs, parseDailyRecommendSong(item, hashPriority))
	}
	return songs
}

// RecommendedSongs 映射卡片推荐歌曲。
func RecommendedSongs(items []kgapi.Record, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		songs = append(songs, parseRecommendedSong(item, hashPriority))
	}
	return songs
}

// PlaylistSongs 映射歌单歌曲，并跳过没有歌名的屏蔽占位条目。
func PlaylistSongs(items []kgapi.Record, hashPriority []string) []dto.Song {
	songs := make([]dto.Song, 0, len(items))
	for _, item := range items {
		song := parsePlaylistSong(item, hashPriority)
		if song.Name == "" {
			continue
		}
		songs = append(songs, song)
	}
	return songs
}

func searchSongName(item kgapi.Record) string {
	if name := kgapi.Text(item["SongName"]); name != "" {
		return name
	}
	return joinSongNameSuffix(kgapi.Text(item["OriSongName"]), kgapi.Text(item["Suffix"]))
}

func lyricSongName(item kgapi.Record) string {
	trans := kgapi.RecordOf(item["trans_param"])
	return joinSongNameSuffix(kgapi.Text(item["SongName"]), kgapi.Text(trans["songname_suffix"]))
}

func joinSongNameSuffix(name string, suffix string) string {
	if name == "" || suffix == "" {
		return name
	}
	return name + " " + suffix
}

func secondsDuration(value any) int64 {
	seconds := kgapi.Int64(value)
	if seconds <= 0 {
		return 0
	}
	return seconds * 1000
}

func parseFlatSong(item kgapi.Record, idKey string, name string, hashPriority []string) dto.Song {
	return dto.Song{
		ID:         kgapi.Int64(item[idKey]),
		Name:       name,
		Artist:     flatArtist(item),
		Album:      kgapi.Text(item["album_name"]),
		CoverURL:   flatCover(item),
		Duration:   flatDuration(item),
		Hash:       play.Resolve(item, hashPriority),
		Quality:    flatQuality(item),
		IsOriginal: isOriginal(item),
	}
}

func parseLatestSong(item kgapi.Record, hashPriority []string) dto.Song {
	return parseFlatSong(item, "album_audio_id", kgapi.Text(item["songname"]), hashPriority)
}

func parseDailyRecommendSong(item kgapi.Record, hashPriority []string) dto.Song {
	name := kgapi.Text(item["songname"])
	if name == "" {
		name = kgapi.Text(item["ori_audio_name"])
	}
	return parseFlatSong(item, "mixsongid", name, hashPriority)
}

func parseRecommendedSong(item kgapi.Record, hashPriority []string) dto.Song {
	audioInfo := kgapi.RecordOf(item["audio_info"])
	albumInfo := kgapi.RecordOf(item["album_info"])
	trans := kgapi.RecordOf(audioInfo["trans_param"])

	duration := kgapi.Int64(audioInfo["timelength"])
	if duration <= 0 {
		duration = flatDuration(item)
	}

	return dto.Song{
		ID:         kgapi.Int64(item["mixsongid"]),
		Name:       joinSongNameSuffix(kgapi.Text(item["ori_audio_name"]), kgapi.Text(trans["songname_suffix"])),
		Artist:     flatArtist(item),
		Album:      kgapi.Text(albumInfo["album_name"]),
		CoverURL:   recommendedCover(item, albumInfo, audioInfo),
		Duration:   duration,
		Hash:       play.Resolve(audioInfo, hashPriority),
		Quality:    flatQuality(audioInfo),
		IsOriginal: isOriginal(audioInfo),
	}
}

func recommendedCover(item kgapi.Record, albumInfo kgapi.Record, audioInfo kgapi.Record) string {
	if cover := kgapi.Text(albumInfo["sizable_cover"]); cover != "" {
		return imageurl.Normalize(cover)
	}
	for _, key := range []string{"cover_pic", "pic"} {
		if cover := kgapi.Text(item[key]); cover != "" {
			return imageurl.Normalize(cover)
		}
	}
	trans := kgapi.RecordOf(audioInfo["trans_param"])
	return imageurl.Normalize(kgapi.Text(trans["union_cover"]))
}

func flatArtist(item kgapi.Record) string {
	if names := collectNames(item, "authors", "author_name"); len(names) > 0 {
		return strings.Join(names, "/")
	}
	if names := collectNames(item, "singerinfo", "name"); len(names) > 0 {
		return strings.Join(names, "/")
	}
	return kgapi.Text(item["author_name"])
}

func collectNames(item kgapi.Record, field string, key string) []string {
	records := kgapi.Records(item[field])
	names := make([]string, 0, len(records))
	for _, record := range records {
		if name := kgapi.Text(record[key]); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func flatCover(item kgapi.Record) string {
	for _, key := range []string{"album_sizable_cover", "sizable_cover"} {
		if cover := kgapi.Text(item[key]); cover != "" {
			return imageurl.Normalize(cover)
		}
	}
	trans := kgapi.RecordOf(item["trans_param"])
	return imageurl.Normalize(kgapi.Text(trans["union_cover"]))
}

func flatDuration(item kgapi.Record) int64 {
	if duration := kgapi.Int64(item["timelength"]); duration > 0 {
		return duration
	}
	if seconds, ok := item["time_length"].(float64); ok && seconds > 0 {
		return int64(seconds * 1000)
	}
	return 0
}

func parsePlaylistSong(item kgapi.Record, hashPriority []string) dto.Song {
	artist, name := splitPlaylistSongName(kgapi.Text(item["name"]))
	if singer := kgapi.Text(kgapi.FirstRecord(item["singerinfo"])["name"]); singer != "" {
		artist = singer
	}

	albumInfo := kgapi.RecordOf(item["albuminfo"])
	cover := kgapi.Text(item["cover"])
	if cover == "" {
		trans := kgapi.RecordOf(item["trans_param"])
		cover = kgapi.Text(trans["union_cover"])
	}

	return dto.Song{
		ID:         kgapi.Int64(item["mixsongid"]),
		Name:       name,
		Artist:     artist,
		Album:      kgapi.Text(albumInfo["name"]),
		CoverURL:   imageurl.Normalize(cover),
		Duration:   kgapi.Int64(item["timelen"]),
		Hash:       play.Resolve(item, hashPriority),
		Quality:    playlistQuality(item),
		IsOriginal: isOriginal(item),
	}
}

func splitPlaylistSongName(full string) (string, string) {
	artist, name, ok := strings.Cut(strings.TrimSpace(full), " - ")
	if !ok {
		return "", strings.TrimSpace(full)
	}
	return strings.TrimSpace(artist), strings.TrimSpace(name)
}

func isOriginal(item kgapi.Record) bool {
	if kgapi.Flag(item["is_original"]) {
		return true
	}
	trans := kgapi.RecordOf(item["trans_param"])
	return kgapi.Flag(trans["is_original"])
}

func searchQuality(item kgapi.Record) string {
	if kgapi.RecordOf(item["Res"]) != nil && kgapi.Int64(item["ResDuration"]) > 0 {
		return qualityHiRes
	}
	if kgapi.Text(item["ResFileHash"]) != "" {
		return qualityHiRes
	}

	if sq := kgapi.RecordOf(item["SQ"]); sq != nil && kgapi.Text(sq["Hash"]) != "" {
		return qualityLossless
	}
	if kgapi.Text(item["SQFileHash"]) != "" && kgapi.Text(item["SQExtName"]) == "flac" {
		return qualityLossless
	}

	if hq := kgapi.RecordOf(item["HQ"]); hq != nil && kgapi.Text(hq["Hash"]) != "" {
		return qualityHQ
	}
	if kgapi.Text(item["HQFileHash"]) != "" {
		if bitrate, ok := item["HQBitrate"].(float64); ok && bitrate == 320 {
			return qualityHQ
		}
	}
	return ""
}

func lyricQuality(item kgapi.Record) string {
	levels := []struct {
		key   string
		label string
	}{
		{"Res", qualityHiRes},
		{"SQ", qualityLossless},
		{"HQ", qualityHQ},
	}
	for _, level := range levels {
		record := kgapi.RecordOf(item[level.key])
		if record != nil && kgapi.Text(record["Hash"]) != "" {
			return level.label
		}
	}
	return ""
}

func flatQuality(item kgapi.Record) string {
	switch {
	case kgapi.Text(item["hash_high"]) != "":
		return qualityHiRes
	case kgapi.Text(item["hash_flac"]) != "":
		return qualityLossless
	case kgapi.Text(item["hash_320"]) != "":
		return qualityHQ
	default:
		return ""
	}
}

func playlistQuality(item kgapi.Record) string {
	maxBitrate := 0
	for _, goods := range kgapi.Records(item["relate_goods"]) {
		if bitrate := kgapi.Int(goods["bitrate"]); bitrate > maxBitrate {
			maxBitrate = bitrate
		}
	}
	switch {
	case maxBitrate >= 1400:
		return qualityHiRes
	case maxBitrate >= 900:
		return qualityLossless
	case maxBitrate >= 320:
		return qualityHQ
	default:
		return ""
	}
}
