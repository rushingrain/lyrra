package search

import "kuwave/internal/dto"

// Lyric 是搜索结果的歌词条目 DTO，json tag 与前端 search.Lyric 对齐。
type Lyric struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

// Result 是搜索服务的稳定响应 DTO。
// 普通搜索只填充其中一个分类；综合搜索按上游分组填充多个分类。
type Result struct {
	Songs     []dto.Song     `json:"songs"`
	Playlists []dto.Playlist `json:"playlists"`
	Albums    []dto.Album    `json:"albums"`
	Artists   []dto.Artist   `json:"artists"`
	MVs       []dto.MV       `json:"mvs"`
	Lyrics    []Lyric        `json:"lyrics"`
}
