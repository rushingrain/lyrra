package play

import (
	"kuwave/internal/auth"
	"kuwave/internal/playerstate"
)

// 播放器本地状态的读写（设备设置 + 当前账号的播放队列 / 当前歌曲）。
//
// 数据落在用户配置目录，由 playerstate 负责分类、原子写与去重：
//   - 设备数据（播放模式 / 音量）：登出保留；
//   - 账号数据（播放队列 / 当前歌曲）：文件里带 userId，登出时由 auth.Logout 删除。
//
// 播放地址与播放进度不在持久化范围内：恢复时只带回歌曲信息，播放地址在真正播放时重新解析。
// 这几个方法不碰 SDK client，因此不占用 Service.mu（并发写由 playerstate 自己的锁兜住）。

// GetPlayerState 读取本地保存的播放状态，供前端启动时恢复。
// 未登录时不返回账号数据（播放队列 / 当前歌曲）。
func (s *Service) GetPlayerState() *playerstate.State {
	state := playerstate.Load(auth.LoadSession())
	return &state
}

// SavePlayerDeviceState 保存设备级设置（播放模式 / 音量）。
func (s *Service) SavePlayerDeviceState(device playerstate.DeviceState) error {
	if err := playerstate.SaveDevice(device); err != nil {
		s.log.Warn("failed to persist player device state", "err", err.Error())
		return err
	}
	return nil
}

// SavePlayerAccountState 保存账号级播放数据（播放队列 / 当前歌曲）。
// 未登录时静默忽略：登出后前端可能还在防抖窗口内把旧数据回写进来。
func (s *Service) SavePlayerAccountState(account playerstate.AccountState) error {
	userID := playerstate.SessionUserID(auth.LoadSession())
	if err := playerstate.SaveAccount(userID, account); err != nil {
		s.log.Warn("failed to persist player account state", "err", err.Error())
		return err
	}
	return nil
}
