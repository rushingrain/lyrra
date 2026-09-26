package search

import "kuwave/internal/model"

// Lyric 是搜索结果的歌词条目 DTO，json tag 与前端 search.Lyric 对齐。
type Lyric struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

// SearchResult 是搜索服务的稳定响应 DTO。
// 普通搜索只填充其中一个分类；综合搜索按上游分组填充多个分类。
type SearchResult struct {
	Songs     []model.Song     `json:"songs"`
	Playlists []model.Playlist `json:"playlists"`
	Albums    []model.Album    `json:"albums"`
	Artists   []model.Artist   `json:"artists"`
	MVs       []model.MV       `json:"mvs"`
	Lyrics    []Lyric          `json:"lyrics"`
}
