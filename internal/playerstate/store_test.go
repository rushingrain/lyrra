package playerstate

import (
	"encoding/json"
	"os"
	"testing"

	"lyrra/internal/dto"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return newStoreWithDir(t.TempDir())
}

func TestLoadReturnsDefaultsWithoutFile(t *testing.T) {
	store := newTestStore(t)

	state := store.Load(nil)
	if state.Device.PlayMode != DefaultPlayMode || state.Device.Volume != DefaultVolume {
		t.Fatalf("device = %+v, want 默认值", state.Device)
	}
	if state.Account.CurrentSong != nil || len(state.Account.Playlist) != 0 {
		t.Fatalf("account = %+v, want 空", state.Account)
	}
}

func TestSaveAndLoadDeviceState(t *testing.T) {
	store := newTestStore(t)

	if err := store.SaveDevice(DeviceState{PlayMode: PlayModeShuffle, Volume: 35}); err != nil {
		t.Fatalf("SaveDevice() error = %v", err)
	}
	state := store.Load(nil)
	if state.Device.PlayMode != PlayModeShuffle || state.Device.Volume != 35 {
		t.Fatalf("device = %+v, want shuffle/35", state.Device)
	}
}

func TestSaveDeviceNormalizesInvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		device   DeviceState
		wantMode string
		wantVol  int
	}{
		{
			name:     "未知播放模式回落默认，音量上限 100",
			device:   DeviceState{PlayMode: "viper", Volume: 180},
			wantMode: DefaultPlayMode,
			wantVol:  100,
		},
		{
			name:     "音量下限 0",
			device:   DeviceState{PlayMode: PlayModeOrder, Volume: -5},
			wantMode: PlayModeOrder,
			wantVol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			if err := store.SaveDevice(tt.device); err != nil {
				t.Fatalf("SaveDevice() error = %v", err)
			}
			state := store.Load(nil)
			if state.Device.PlayMode != tt.wantMode || state.Device.Volume != tt.wantVol {
				t.Fatalf("device = %+v, want %s/%d", state.Device, tt.wantMode, tt.wantVol)
			}
		})
	}
}

// TestSaveDeviceSkipsUnchangedContent 验证去重：内容没变时不再写盘。
// 手动把文件改坏后再存同样的内容，文件应保持"改坏"的样子（说明没有重写）。
func TestSaveDeviceSkipsUnchangedContent(t *testing.T) {
	store := newTestStore(t)
	device := DeviceState{PlayMode: PlayModeSingle, Volume: 20}
	if err := store.SaveDevice(device); err != nil {
		t.Fatalf("SaveDevice() error = %v", err)
	}
	if err := os.WriteFile(store.devicePath(), []byte("junk"), 0o600); err != nil {
		t.Fatalf("写入临时内容失败: %v", err)
	}

	if err := store.SaveDevice(device); err != nil {
		t.Fatalf("SaveDevice() 重复调用 error = %v", err)
	}
	payload, err := os.ReadFile(store.devicePath())
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(payload) != "junk" {
		t.Fatalf("内容未变时不应重写文件，实际 %q", payload)
	}

	// 内容变化时必须重新写盘
	if err := store.SaveDevice(DeviceState{PlayMode: PlayModeSingle, Volume: 21}); err != nil {
		t.Fatalf("SaveDevice() 变更后 error = %v", err)
	}
	if state := store.Load(nil); state.Device.Volume != 21 {
		t.Fatalf("device = %+v, want volume 21", state.Device)
	}
}

func testAccount() AccountState {
	return AccountState{
		Playlist: []dto.Song{
			{ID: 1, Name: "第一首", Hash: "hash-1"},
			{ID: 2, Name: "第二首", Hash: "hash-2"},
		},
		CurrentSong: &dto.Song{ID: 2, Name: "第二首", Hash: "hash-2"},
	}
}

func TestSaveAccountOnlyVisibleToSameUser(t *testing.T) {
	store := newTestStore(t)
	if err := store.SaveAccount("1001", testAccount()); err != nil {
		t.Fatalf("SaveAccount() error = %v", err)
	}

	got := store.Load(map[string]string{"userid": "1001"})
	if len(got.Account.Playlist) != 2 {
		t.Fatalf("playlist = %+v, want 2 首", got.Account.Playlist)
	}
	if got.Account.CurrentSong == nil || got.Account.CurrentSong.ID != 2 {
		t.Fatalf("currentSong = %+v, want id 2", got.Account.CurrentSong)
	}

	// 换账号、未登录、userid=0 都读不到账号数据
	for name, session := range map[string]map[string]string{
		"换账号":        {"userid": "1002"},
		"未登录":        nil,
		"userid 为 0": {"userid": "0"},
	} {
		t.Run(name, func(t *testing.T) {
			state := store.Load(session)
			if state.Account.CurrentSong != nil || len(state.Account.Playlist) != 0 {
				t.Fatalf("account = %+v, want 空", state.Account)
			}
		})
	}
}

func TestSaveAccountWithoutUserIsIgnored(t *testing.T) {
	store := newTestStore(t)

	if err := store.SaveAccount("", testAccount()); err != nil {
		t.Fatalf("SaveAccount() error = %v", err)
	}
	if _, err := os.Stat(store.accountPath()); !os.IsNotExist(err) {
		t.Fatalf("未登录时不应创建账号文件，stat err = %v", err)
	}
}

func TestClearAccountKeepsDeviceState(t *testing.T) {
	store := newTestStore(t)
	if err := store.SaveDevice(DeviceState{PlayMode: PlayModeShuffle, Volume: 42}); err != nil {
		t.Fatalf("SaveDevice() error = %v", err)
	}
	if err := store.SaveAccount("1001", testAccount()); err != nil {
		t.Fatalf("SaveAccount() error = %v", err)
	}

	if err := store.ClearAccount(); err != nil {
		t.Fatalf("ClearAccount() error = %v", err)
	}
	if _, err := os.Stat(store.accountPath()); !os.IsNotExist(err) {
		t.Fatalf("账号文件应被删除，stat err = %v", err)
	}
	if state := store.Load(map[string]string{"userid": "1001"}); state.Account.CurrentSong != nil {
		t.Fatalf("清空后仍读到账号数据: %+v", state.Account)
	}
	// 设备数据不受登出影响
	if state := store.Load(nil); state.Device.PlayMode != PlayModeShuffle || state.Device.Volume != 42 {
		t.Fatalf("device = %+v, want shuffle/42", state.Device)
	}
}

// TestClearAccountAllowsRewriteWithSameContent 覆盖登出后的回写：
// 前端防抖窗口内可能把相同内容再发一次，去重缓存必须已被清掉，否则数据会被永久丢弃。
func TestClearAccountAllowsRewriteWithSameContent(t *testing.T) {
	store := newTestStore(t)
	account := testAccount()
	if err := store.SaveAccount("1001", account); err != nil {
		t.Fatalf("SaveAccount() error = %v", err)
	}
	if err := store.ClearAccount(); err != nil {
		t.Fatalf("ClearAccount() error = %v", err)
	}

	if err := store.SaveAccount("1001", account); err != nil {
		t.Fatalf("SaveAccount() error = %v", err)
	}
	if _, err := os.Stat(store.accountPath()); err != nil {
		t.Fatalf("清空后再次保存应写盘: %v", err)
	}
}

func TestLoadToleratesCorruptFiles(t *testing.T) {
	store := newTestStore(t)
	if err := os.MkdirAll(store.dir, 0o700); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	for _, path := range []string{store.devicePath(), store.accountPath()} {
		if err := os.WriteFile(path, []byte(`{"playMode":`), 0o600); err != nil {
			t.Fatalf("写入损坏文件失败: %v", err)
		}
	}

	state := store.Load(map[string]string{"userid": "1001"})
	if state.Device.PlayMode != DefaultPlayMode || state.Device.Volume != DefaultVolume {
		t.Fatalf("device = %+v, want 默认值", state.Device)
	}
	if state.Account.CurrentSong != nil {
		t.Fatalf("account = %+v, want 空", state.Account)
	}
}

func TestSessionUserID(t *testing.T) {
	tests := []struct {
		name    string
		session map[string]string
		want    string
	}{
		{name: "正常", session: map[string]string{"userid": "1001"}, want: "1001"},
		{name: "去空白", session: map[string]string{"userid": " 1001 "}, want: "1001"},
		{name: "userid 为 0 视为未登录", session: map[string]string{"userid": "0"}, want: ""},
		{name: "空会话", session: nil, want: ""},
		{name: "缺字段", session: map[string]string{"token": "t"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SessionUserID(tt.session); got != tt.want {
				t.Fatalf("SessionUserID() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSaveAccountTrimsPlaylist 队列上限：超出部分丢最早的，保留最新的 MaxPlaylistSize 条。
func TestSaveAccountTrimsPlaylist(t *testing.T) {
	store := newTestStore(t)
	playlist := make([]dto.Song, 0, MaxPlaylistSize+20)
	for i := 1; i <= MaxPlaylistSize+20; i++ {
		playlist = append(playlist, dto.Song{ID: int64(i), Name: "song", Hash: "hash"})
	}

	if err := store.SaveAccount("1001", AccountState{
		Playlist:    playlist,
		CurrentSong: &playlist[len(playlist)-1],
	}); err != nil {
		t.Fatalf("SaveAccount() error = %v", err)
	}

	got := store.Load(map[string]string{"userid": "1001"}).Account
	if len(got.Playlist) != MaxPlaylistSize {
		t.Fatalf("playlist 长度 = %d, want %d", len(got.Playlist), MaxPlaylistSize)
	}
	// 前 20 条被丢掉，保留 21..520
	if got.Playlist[0].ID != 21 {
		t.Fatalf("队首 id = %d, want 21", got.Playlist[0].ID)
	}
	if last := got.Playlist[len(got.Playlist)-1].ID; last != int64(MaxPlaylistSize+20) {
		t.Fatalf("队尾 id = %d, want %d", last, MaxPlaylistSize+20)
	}
	if got.CurrentSong == nil || got.CurrentSong.ID != int64(MaxPlaylistSize+20) {
		t.Fatalf("currentSong = %+v, want 最后一条", got.CurrentSong)
	}
}

// TestLoadTrimsOversizedPlaylist 历史文件超过上限时，读取同样按"留最新"裁剪。
func TestLoadTrimsOversizedPlaylist(t *testing.T) {
	store := newTestStore(t)
	if err := os.MkdirAll(store.dir, 0o700); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	oversized := make([]dto.Song, 0, MaxPlaylistSize+5)
	for i := 1; i <= MaxPlaylistSize+5; i++ {
		oversized = append(oversized, dto.Song{ID: int64(i), Name: "song", Hash: "hash"})
	}
	payload, err := json.Marshal(map[string]any{"userId": "1001", "playlist": oversized})
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	if err := os.WriteFile(store.accountPath(), payload, 0o600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	got := store.Load(map[string]string{"userid": "1001"}).Account.Playlist
	if len(got) != MaxPlaylistSize {
		t.Fatalf("playlist 长度 = %d, want %d", len(got), MaxPlaylistSize)
	}
	if got[0].ID != 6 {
		t.Fatalf("队首 id = %d, want 6（前 5 条被丢）", got[0].ID)
	}
}
