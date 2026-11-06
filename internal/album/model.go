package album

import "lyrra/internal/dto"

// SongsResult 是专辑歌曲分页结果。
type SongsResult struct {
	Songs []dto.Song `json:"songs"`
	// Total 专辑总歌曲数
	Total int64 `json:"total"`
	// HasMore 是否还有下一页
	HasMore bool `json:"hasMore"`
}
