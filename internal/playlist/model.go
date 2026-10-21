package playlist

import "lyrra/internal/dto"

// SongsResult 是歌单歌曲分页结果。
//
// Info 由 PlaylistTrackAll 的 list_info 解析；该接口不返回播放次数，
// 前端用首页传入的歌单信息补齐。
type SongsResult struct {
	Info  dto.Playlist `json:"info"`
	Songs []dto.Song   `json:"songs"`
	// Total 歌单总歌曲数
	Total int64 `json:"total"`
	// HasMore 是否还有下一页
	HasMore bool `json:"hasMore"`
}
