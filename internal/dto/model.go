// Package dto 存放跨服务复用的音乐实体 DTO。
package dto

// Song 是音乐单曲 DTO，JSON 字段与前端 types/index.ts 的 Song 类型对齐。
type Song struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album,omitempty"`
	CoverURL string `json:"coverUrl"`
	Duration int64  `json:"duration,omitempty"`
	// Hash 播放用文件 hash：按播放品质偏好（默认标准）从高到低解析出的第一个可用值，
	// 为空表示该条目没有任何可用音质，前端需给出提示
	Hash       string   `json:"hash,omitempty"`
	Quality    string   `json:"quality,omitempty"`
	IsOriginal bool     `json:"isOriginal"`
	Tags       []string `json:"tags,omitempty"`
	// MatchedLyric 歌词搜索时命中的歌词片段，其它来源不返回
	MatchedLyric string `json:"matchedLyric,omitempty"`
}

// Playlist 是歌单 DTO。
type Playlist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	CoverURL    string `json:"coverUrl"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
	PlayCount   int64  `json:"playCount"`
	SongCount   int64  `json:"songCount,omitempty"`
}

// Album 是专辑 DTO。
type Album struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	CoverURL string `json:"coverUrl"`
	Date     string `json:"date"`
}

// Artist 是歌手 DTO。
type Artist struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	SongCount int64  `json:"songCount"`
	Genre     string `json:"genre"`
	Fans      int64  `json:"fans"`
}

// MV 是 MV DTO。
type MV struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	CoverURL string `json:"coverUrl"`
}

type Banner struct {
	ID       string `json:"id"`
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
	LinkURL  string `json:"linkUrl"`
}
