package playerstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"lyrra/internal/appdir"
	"lyrra/internal/atomicfile"
)

// 存储位置与 session.json / vip.json 同目录（应用数据目录，见 internal/appdir）。
const (
	deviceFileName  = "player-device.json"
	accountFileName = "player-account.json"
)

// Store 负责播放状态文件的读写。
//
// 两个文件分开存是为了让"账号数据随登出删除、设备数据保留"这条规则落在文件粒度上。
// 写入统一走"同目录临时文件 + rename"的原子写；并用最近一次写入的内容做去重，
// 避免前端抖动（音量拖动、连续切歌）把同样的内容反复落盘。
type Store struct {
	dir string

	mu          sync.Mutex
	lastDevice  string
	lastAccount string
}

// NewStore 返回使用用户配置目录的 Store。
func NewStore() *Store {
	return newStoreWithDir(appdir.Dir())
}

// newStoreWithDir 指定目录构造，便于单测写临时目录。
func newStoreWithDir(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) devicePath() string {
	return filepath.Join(s.dir, deviceFileName)
}

func (s *Store) accountPath() string {
	return filepath.Join(s.dir, accountFileName)
}

// Load 读取播放状态。
//
// 设备数据总是返回（文件缺失或损坏时用默认值）；账号数据只在已登录、
// 且文件里的 userId 与当前账号一致时返回，否则返回空数据。
// 恢复失败不应该阻塞应用启动，因此这里不返回错误。
func (s *Store) Load(session map[string]string) State {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := State{
		Device: DeviceState{PlayMode: DefaultPlayMode, Volume: DefaultVolume},
		Account: AccountState{
			Playlist: normalizePlaylist(nil),
		},
	}

	if payload, err := os.ReadFile(s.devicePath()); err == nil {
		var device DeviceState
		if json.Unmarshal(payload, &device) == nil {
			state.Device = normalizeDevice(device)
		}
	}
	// 记成"已落盘的内容"：恢复后前端即使回写同一份设置，也不会再写一次磁盘。
	if payload, err := json.Marshal(state.Device); err == nil {
		s.lastDevice = string(payload)
	}

	userID := SessionUserID(session)
	if userID == "" {
		return state
	}
	payload, err := os.ReadFile(s.accountPath())
	if err != nil {
		return state
	}
	var file accountFile
	if json.Unmarshal(payload, &file) != nil {
		return state
	}
	// 空 userId 视为历史数据，宽容接受；有值则必须与当前账号一致。
	if file.UserID != "" && file.UserID != userID {
		return state
	}
	state.Account = AccountState{
		Playlist:    normalizePlaylist(file.Playlist),
		CurrentSong: file.CurrentSong,
	}
	if payload, err := json.Marshal(file); err == nil {
		s.lastAccount = string(payload)
	}
	return state
}

// SaveDevice 保存设备级设置。内容与上次写入一致时直接跳过。
func (s *Store) SaveDevice(device DeviceState) error {
	payload, err := json.Marshal(normalizeDevice(device))
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if string(payload) == s.lastDevice {
		return nil
	}
	if err := atomicfile.Write(s.devicePath(), payload); err != nil {
		return err
	}
	s.lastDevice = string(payload)
	return nil
}

// SaveAccount 保存账号级播放数据。未登录（userID 为空）时直接忽略：
// 登出后前端仍可能在防抖窗口内回写，这里必须挡住。
func (s *Store) SaveAccount(userID string, account AccountState) error {
	if userID == "" {
		return nil
	}
	payload, err := json.Marshal(accountFile{
		AccountState: AccountState{
			Playlist:    normalizePlaylist(account.Playlist),
			CurrentSong: account.CurrentSong,
		},
		UserID: userID,
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if string(payload) == s.lastAccount {
		return nil
	}
	if err := atomicfile.Write(s.accountPath(), payload); err != nil {
		return err
	}
	s.lastAccount = string(payload)
	return nil
}

// ClearAccount 删除账号级播放数据，登出时调用；设备数据不动。
func (s *Store) ClearAccount() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 清掉去重缓存，避免登出后回写被误判成"内容没变"。
	s.lastAccount = ""
	if err := os.Remove(s.accountPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// 默认 Store：play 服务读写、auth 登出清理共用同一份路径与去重缓存。
var defaultStore = NewStore()

// Load 读取播放状态（默认 Store）。
func Load(session map[string]string) State {
	return defaultStore.Load(session)
}

// SaveDevice 保存设备级设置（默认 Store）。
func SaveDevice(device DeviceState) error {
	return defaultStore.SaveDevice(device)
}

// SaveAccount 保存账号级播放数据（默认 Store）。
func SaveAccount(userID string, account AccountState) error {
	return defaultStore.SaveAccount(userID, account)
}

// ClearAccount 删除账号级播放数据（默认 Store），登出时调用。
func ClearAccount() error {
	return defaultStore.ClearAccount()
}
