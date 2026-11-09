package search

import "lyrra/internal/dto"

// Result 是搜索服务的稳定响应 DTO。
// 普通搜索只填充其中一个分类；综合搜索按上游分组填充多个分类。
type Result struct {
	Songs     []dto.Song     `json:"songs"`
	Playlists []dto.Playlist `json:"playlists"`
	Albums    []dto.Album    `json:"albums"`
	Artists   []dto.Artist   `json:"artists"`
	MVs       []dto.MV       `json:"mvs"`
}
