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
	// Hash 播放用文件 hash，只用于定位歌曲：实际返回的文件由 play.GetSongPlayURL 的
	// quality 参数决定，因此这里取候选档位里第一个非空值（默认标准档缺档时向上兜底），
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
	ID int64 `json:"id"`
	// GlobalCollectionID 歌单的上游全局 ID，用于请求歌单歌曲
	GlobalCollectionID string   `json:"globalCollectionId"`
	Name               string   `json:"name"`
	CoverURL           string   `json:"coverUrl"`
	Description        string   `json:"description"`
	Creator            string   `json:"creator"`
	CreatorAvatar      string   `json:"creatorAvatar,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	// PublishTime 创建时间，格式由上游决定（如 "2023-03-31 00:00:00"）
	PublishTime string `json:"publishTime,omitempty"`
	// ContainedSong 歌单搜索命中的歌曲名，仅普通歌单搜索返回
	ContainedSong string `json:"containedSong,omitempty"`
	PlayCount     int64  `json:"playCount"`
	SongCount     int64  `json:"songCount,omitempty"`
}

// Album 是专辑 DTO。
type Album struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Artist      string `json:"artist"`
	CoverURL    string `json:"coverUrl"`
	Date        string `json:"date"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Language    string `json:"language,omitempty"`
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
