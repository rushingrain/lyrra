package playerstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// 存储位置与 session.json / vip.json 同目录（os.UserConfigDir()/kuwave/）。
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
	return newStoreWithDir(configDir())
}

// newStoreWithDir 指定目录构造，便于单测写临时目录。
func newStoreWithDir(dir string) *Store {
	return &Store{dir: dir}
}

// configDir 返回应用的数据目录（macOS: ~/Library/Application Support/kuwave/）。
// 与 auth/session.go、vip/store.go 保持同一约定，不依赖当前工作目录。
func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "kuwave")
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
	if err := writeFileAtomic(s.devicePath(), payload); err != nil {
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
	if err := writeFileAtomic(s.accountPath(), payload); err != nil {
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

// writeFileAtomic 先写同目录临时文件再 rename，避免进程中断把主文件截断成半截 JSON。
// 权限与 session.json 一致（目录 0700、文件 0600）。
func writeFileAtomic(path string, payload []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	// 先落盘再改名，避免 rename 成功但内容还在页缓存里
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
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
