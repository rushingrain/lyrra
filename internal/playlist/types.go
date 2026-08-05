package playlist

// 首页官方歌单 DTO，json tag 与 frontend/src/types/index.ts 的 Playlist 对齐。
// 注：/top/playlist 接口不返回歌单内歌曲数，前端 Playlist.songCount 已设为可选。

// HomePlaylist 首页官方歌单。
type HomePlaylist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CoverURL    string `json:"coverUrl"`
	Description string `json:"description"`
	PlayCount   int    `json:"playCount"`
}
