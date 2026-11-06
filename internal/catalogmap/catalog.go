package catalogmap

import (
	"fmt"
	"strings"

	"lyrra/internal/dto"
	"lyrra/internal/imageurl"
	"lyrra/internal/kgapi"
)

// SearchPlaylists 映射搜索结果中的歌单。
func SearchPlaylists(items []kgapi.Record) []dto.Playlist {
	return searchPlaylists(items, false)
}

// NormalSearchPlaylists 映射普通歌单搜索，并补充命中的歌曲名。
func NormalSearchPlaylists(items []kgapi.Record) []dto.Playlist {
	return searchPlaylists(items, true)
}

func searchPlaylists(items []kgapi.Record, includeContainedSong bool) []dto.Playlist {
	playlists := make([]dto.Playlist, 0, len(items))
	for _, item := range items {
		playlist := dto.Playlist{
			ID:                 kgapi.Int64(item["specialid"]),
			GlobalCollectionID: kgapi.Text(item["gid"]),
			Name:               kgapi.Text(item["specialname"]),
			Creator:            kgapi.Text(item["nickname"]),
			CoverURL:           imageurl.Normalize(kgapi.Text(item["img"])),
			Description:        kgapi.Text(item["intro"]),
			PublishTime:        kgapi.Text(item["publish_time"]),
			PlayCount:          kgapi.Int64(item["total_play_count"]),
			SongCount:          kgapi.Int64(item["song_count"]),
		}
		if includeContainedSong {
			playlist.ContainedSong = kgapi.Text(item["contain"])
		}
		playlists = append(playlists, playlist)
	}
	return playlists
}

// HotPlaylists 映射首页官方歌单。
func HotPlaylists(items []kgapi.Record) []dto.Playlist {
	playlists := make([]dto.Playlist, 0, len(items))
	for _, item := range items {
		playlists = append(playlists, dto.Playlist{
			ID:                 kgapi.Int64(item["specialid"]),
			GlobalCollectionID: kgapi.Text(item["global_collection_id"]),
			Name:               kgapi.Text(item["specialname"]),
			Creator:            kgapi.Text(item["nickname"]),
			CreatorAvatar: imageurl.Normalize(kgapi.FirstNonEmpty(
				kgapi.Text(item["user_avatar"]),
				kgapi.Text(item["pic"]),
			)),
			Tags:        tagNames(item["tags"]),
			PublishTime: kgapi.Text(item["publishtime"]),
			CoverURL:    imageurl.Normalize(kgapi.Text(item["imgurl"])),
			Description: kgapi.Text(item["intro"]),
			PlayCount:   kgapi.Int64(item["play_count"]),
			SongCount:   kgapi.Int64(item["song_count"]),
		})
	}
	return playlists
}

// PlaylistInfo 映射 PlaylistTrackAll 的 list_info。
func PlaylistInfo(data kgapi.Record) dto.Playlist {
	info := kgapi.RecordOf(data["list_info"])
	return dto.Playlist{
		ID:                 kgapi.Int64(info["specialid"]),
		GlobalCollectionID: kgapi.Text(info["global_collection_id"]),
		Name:               kgapi.Text(info["name"]),
		CoverURL:           imageurl.Normalize(kgapi.Text(info["pic"])),
		Description:        kgapi.Text(info["intro"]),
		Creator:            kgapi.Text(info["list_create_username"]),
		CreatorAvatar:      imageurl.Normalize(kgapi.Text(info["create_user_pic"])),
		Tags:               tagNames(info["musiclib_tags"]),
		PublishTime:        kgapi.Text(info["publish_date"]),
		SongCount:          kgapi.Int64(info["count"]),
	}
}

// Albums 映射搜索结果中的专辑。
func Albums(items []kgapi.Record) []dto.Album {
	albums := make([]dto.Album, 0, len(items))
	for _, item := range items {
		albums = append(albums, dto.Album{
			ID:       kgapi.Int64(item["albumid"]),
			Name:     kgapi.Text(item["albumname"]),
			Artist:   searchArtist(item, "singer", "singers"),
			CoverURL: kgapi.Text(item["img"]),
			Date:     strings.SplitN(kgapi.Text(item["publish_time"]), " ", 2)[0],
		})
	}
	return albums
}

// Artists 映射搜索结果中的歌手。
func Artists(items []kgapi.Record) []dto.Artist {
	artists := make([]dto.Artist, 0, len(items))
	for _, item := range items {
		artists = append(artists, dto.Artist{
			ID:        kgapi.Int64(item["AuthorId"]),
			Name:      kgapi.Text(item["AuthorName"]),
			AvatarURL: kgapi.Text(item["Avatar"]),
			SongCount: kgapi.Int64(item["AudioCount"]),
			Genre:     kgapi.Text(item["Genre"]),
			Fans:      kgapi.Int64(item["FansNum"]),
		})
	}
	return artists
}

// MVs 映射综合搜索返回的 MV。
func MVs(items []kgapi.Record) []dto.MV {
	mvs := make([]dto.MV, 0, len(items))
	for _, item := range items {
		mvs = append(mvs, dto.MV{
			ID:       kgapi.Int64(item["MvID"]),
			Name:     kgapi.Text(item["MvName"]),
			Artist:   searchArtist(item, "SingerName", "Singers"),
			CoverURL: kgapi.Text(item["Pic"]),
		})
	}
	return mvs
}

// NormalSearchMVs 映射普通搜索 MV，并补全只返回文件名的封面地址。
func NormalSearchMVs(items []kgapi.Record) []dto.MV {
	mvs := MVs(items)
	for index := range mvs {
		mvs[index].CoverURL = normalizeMVPicFilename(mvs[index].CoverURL)
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
	for _, char := range date {
		if char < '0' || char > '9' {
			return value
		}
	}

	return imageurl.Normalize(fmt.Sprintf("http://imge.kugou.com/mvhdpic/400/%s/%s", date, filename))
}

func searchArtist(item kgapi.Record, nameKey string, singersKey string) string {
	if name := kgapi.Text(item[nameKey]); name != "" {
		return name
	}
	singer := kgapi.FirstRecord(item[singersKey])
	return kgapi.Text(singer["name"])
}

func splitTags(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(char rune) bool {
		return char == '|' || char == ','
	})
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			tags = append(tags, part)
		}
	}
	return tags
}

func tagNames(value any) []string {
	records := kgapi.Records(value)
	names := make([]string, 0, len(records))
	for _, record := range records {
		if name := kgapi.Text(record["tag_name"]); name != "" {
			names = append(names, name)
		}
	}
	return names
}
