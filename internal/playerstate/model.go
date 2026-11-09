// Package playerstate 负责播放器本地状态的持久化。
//
// 数据分两类存放：
//   - 设备数据（播放模式、音量）：与账号无关，登出后保留；
//   - 账号数据（播放队列、当前歌曲）：随登出删除，文件里带 userId 用于校验归属。
//
// 播放进度不持久化；播放地址（带签名的 url）也不落盘，恢复后由播放接口重新解析。
package playerstate

import (
	"strings"

	"lyrra/internal/dto"
)

// 播放模式取值：与前端 playerStore 的 PlayMode 保持一致。
const (
	PlayModeListLoop = "list-loop"
	PlayModeOrder    = "order"
	PlayModeSingle   = "single"
	PlayModeShuffle  = "shuffle"
)

// 默认值：文件缺失或字段非法时回落到这里，与前端默认值保持一致。
const (
	DefaultPlayMode = PlayModeListLoop
	DefaultVolume   = 60
)

// MaxPlaylistSize 播放队列上限。
// 前端在入列时就会拒绝超限（提示用户先清空），这里只是兜底：
// 读取/写入超出上限的文件（例如旧版本留下的）时保留最新的这些条目。
// 与前端 playerStore 的 MAX_PLAYLIST_SIZE 保持一致。
const MaxPlaylistSize = 500

// validPlayModes 播放模式白名单，避免把前端传错的值写进磁盘。
var validPlayModes = map[string]struct{}{
	PlayModeListLoop: {},
	PlayModeOrder:    {},
	PlayModeSingle:   {},
	PlayModeShuffle:  {},
}

// DeviceState 设备级播放设置：与账号无关，登出后保留。
type DeviceState struct {
	PlayMode string `json:"playMode"`
	Volume   int    `json:"volume"`
}

// AccountState 账号级播放数据：登出即删除。
// 只有歌曲信息（含播放用的 hash），不含播放地址与播放进度。
type AccountState struct {
	Playlist    []dto.Song `json:"playlist"`
	CurrentSong *dto.Song  `json:"currentSong,omitempty"`
}

// State 是启动时一次性读取的播放状态。
type State struct {
	Device  DeviceState  `json:"device"`
	Account AccountState `json:"account"`
}

// accountFile 是账号数据的落盘结构：多带一个 userId，
// 读取时用来确认这份数据属于当前登录账号（换账号后不会误用上一个人的队列）。
type accountFile struct {
	AccountState
	UserID string `json:"userId,omitempty"`
}

// normalizeDevice 规整设备设置：播放模式必须在白名单内，音量夹到 0-100。
func normalizeDevice(state DeviceState) DeviceState {
	if _, ok := validPlayModes[state.PlayMode]; !ok {
		state.PlayMode = DefaultPlayMode
	}
	state.Volume = clampVolume(state.Volume)
	return state
}

func clampVolume(volume int) int {
	if volume < 0 {
		return 0
	}
	if volume > 100 {
		return 100
	}
	return volume
}

// SessionUserID 从会话里取当前登录用户 id；未登录（或 userid 为 0）时返回空串。
// 会话结构见 auth.LoadSession。
func SessionUserID(session map[string]string) string {
	userID := strings.TrimSpace(session["userid"])
	if userID == "0" {
		return ""
	}
	return userID
}

// normalizePlaylist 规整播放队列：保证序列化成 [] 而不是 null（前端少一次空值判断），
// 并裁剪到上限——队尾是最新的，所以保留末尾 MaxPlaylistSize 条。
func normalizePlaylist(songs []dto.Song) []dto.Song {
	if songs == nil {
		return []dto.Song{}
	}
	if len(songs) > MaxPlaylistSize {
		return songs[len(songs)-MaxPlaylistSize:]
	}
	return songs
}
