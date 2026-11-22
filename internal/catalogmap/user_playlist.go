package catalogmap

import (
	"strings"

	"lyrra/internal/dto"
	"lyrra/internal/imageurl"
	"lyrra/internal/kgapi"
)

// UserPlaylist 映射账号歌单；收藏项使用原歌单的全局 ID 读取歌曲。
func UserPlaylist(item kgapi.Record) dto.UserPlaylist {
	typeValue, ok := kgapi.IntOK(item["type"])
	if !ok {
		typeValue = -1
	}
	isDefault, ok := kgapi.IntOK(item["is_def"])
	// 上游只为特殊歌单返回 is_def，普通歌单省略该字段。
	if item["is_def"] == nil {
		isDefault = 0
	} else if !ok {
		isDefault = -1
	}
	collectionID := kgapi.Scalar(item["global_collection_id"])
	if typeValue == 1 {
		if sourceID := kgapi.Scalar(item["list_create_gid"]); sourceID != "" && sourceID != "0" {
			collectionID = sourceID
		}
	}
	return dto.UserPlaylist{
		ID:                 kgapi.Scalar(item["listid"]),
		GlobalCollectionID: collectionID,
		Name:               kgapi.Text(item["name"]),
		SongCount:          kgapi.Int(item["count"]),
		CoverURL:           imageurl.Normalize(kgapi.Text(item["pic"])),
		Description:        kgapi.Text(item["intro"]),
		Tags:               userPlaylistTags(item["tags"]),
		IsDefault:          isDefault,
		Type:               typeValue,
	}
}

func userPlaylistTags(value any) []string {
	if text, ok := value.(string); ok {
		tags := make([]string, 0)
		for _, tag := range strings.Split(text, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				tags = append(tags, tag)
			}
		}
		return tags
	}
	if items, ok := value.([]any); ok && len(items) > 0 {
		if _, stringsOnly := items[0].(string); stringsOnly {
			tags := make([]string, 0, len(items))
			for _, item := range items {
				if tag := kgapi.Text(item); tag != "" {
					tags = append(tags, tag)
				}
			}
			return tags
		}
	}
	return tagNames(value)
}

// UserPlaylistWithInfo 在账号列表缺少编辑字段时，使用公开详情补充。
func UserPlaylistWithInfo(item kgapi.Record, info dto.Playlist) dto.UserPlaylist {
	playlist := UserPlaylist(item)
	if _, exists := item["intro"]; !exists {
		playlist.Description = info.Description
	}
	if _, exists := item["tags"]; !exists {
		playlist.Tags = info.Tags
	}
	return playlist
}
