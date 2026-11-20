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
type UserPlaylist struct {
	ID string `json:"id"`
	// GlobalCollectionID 用于获取歌单歌曲，与用户歌单的 listid 区分。
	GlobalCollectionID string `json:"globalCollectionId"`
	Name               string `json:"name"`
	SongCount          int    `json:"songCount"`
	CoverURL           string `json:"coverUrl"`
	// 歌单类型：1=默认收藏，2=我喜欢，0=普通歌单
	IsDefault int `json:"isDefault"`
	// 归属：0=创建的歌单，1=收藏的歌单
	Type int `json:"type"`
}
