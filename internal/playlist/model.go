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

// UserPlaylist 用户歌单 DTO（侧边栏展示）。
// json tag 与前端 frontend/src/types/index.ts 的 UserPlaylist 类型对齐。
type UserPlaylist = dto.UserPlaylist

// PlaylistEdit 只包含用户可编辑的字段；版本号、排序和归属由后端读取。
type PlaylistEdit struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	CoverData   string   `json:"coverData"`
}
