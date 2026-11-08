package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lyrra/internal/appdir"
	"lyrra/internal/atomicfile"
)

// 会话持久化：只保存关键的登录态 cookie 字段。
// 存储位置为用户配置目录（macOS: ~/Library/Application Support/lyrra/、
// Windows: %AppData%\lyrra\、Linux: ~/.config/lyrra/），
// 不依赖当前工作目录（桌面应用 cwd 不可靠）。

// 登录有效期从扫码确认成功时开始计算，不做活动续期。
const sessionTTL = 5 * 24 * time.Hour

// sessionFile 是 session.json 的持久化结构。
type sessionFile struct {
	Token     string `json:"token"`
	UserID    string `json:"userid"`
	DFID      string `json:"dfid"`
	VipToken  string `json:"vip_token"`
	VipType   string `json:"vip_type"`
	ExpiresAt int64  `json:"expiresAt"`
}

func sessionPath() string {
	return filepath.Join(appdir.Dir(), "session.json")
}

// saveSession 保存会话，登录成功后调用。
// 固定写入 token / userid / dfid / vip_token / vip_type 五个登录态 cookie 字段。
// 非 VIP 账号的 vip_token / vip_type 可能为空，但保持一致的结构便于后续扩展。
func saveSession(cookie map[string]string) {
	data, err := json.Marshal(newSessionFile(cookie, time.Now()))
	if err != nil {
		return
	}
	_ = atomicfile.Write(sessionPath(), data)
}

func newSessionFile(cookie map[string]string, now time.Time) sessionFile {
	return sessionFile{
		Token:     cookie["token"],
		UserID:    normalizeUserIDString(cookie["userid"]),
		DFID:      cookie["dfid"],
		VipToken:  cookie["vip_token"],
		VipType:   cookie["vip_type"],
		ExpiresAt: now.Add(sessionTTL).UnixMilli(),
	}
}

// LoginCookie 选取 token/userid/dfid 三个有效字段，保持原值。
func LoginCookie(session map[string]string) map[string]string {
	return selectLoginCookie(session, false)
}

// TrimmedLoginCookie 选取 token/userid/dfid 三个有效字段并去掉首尾空白。
func TrimmedLoginCookie(session map[string]string) map[string]string {
	return selectLoginCookie(session, true)
}

func selectLoginCookie(session map[string]string, trim bool) map[string]string {
	cookie := make(map[string]string, 3)
	for _, key := range []string{"token", "userid", "dfid"} {
		value := session[key]
		if trim {
			value = strings.TrimSpace(value)
		}
		if value != "" {
			cookie[key] = value
		}
	}
	return cookie
}

// LoadSession 恢复会话。无会话时返回 nil。
// 除 Service 外，需使用当前登录态的服务也通过该方法读取会话。
func LoadSession() map[string]string {
	session, _, _ := loadSession()
	return session
}

// loadSession 返回有效会话及其过期时间。expired 用于区分“从未登录”和“会话已过期”，
// 便于调用方清理内存中的登录态和账号级数据。
func loadSession() (map[string]string, int64, bool) {
	path := sessionPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, false
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, false
	}
	stored, err := decodeSession(data, info.ModTime())
	if err != nil {
		return nil, 0, false
	}
	if sessionExpired(stored.ExpiresAt, time.Now()) {
		clearSession()
		return nil, 0, true
	}
	return sessionCookie(stored), stored.ExpiresAt, false
}

func sessionExpired(expiresAt int64, now time.Time) bool {
	return !now.Before(time.UnixMilli(expiresAt))
}

func decodeSession(data []byte, fallbackTime time.Time) (sessionFile, error) {
	var stored sessionFile
	if err := json.Unmarshal(data, &stored); err != nil {
		return sessionFile{}, err
	}
	// 兼容没有过期时间的旧会话：以文件最后写入时间作为登录时间。
	if stored.ExpiresAt <= 0 {
		stored.ExpiresAt = fallbackTime.Add(sessionTTL).UnixMilli()
	}
	return stored, nil
}

func sessionCookie(stored sessionFile) map[string]string {
	// 兼容历史会话：早期扫码登录可能把 userid 存成科学计数法
	// （如 "8.63006722e+08"），加载时统一规整为十进制整数。
	return map[string]string{
		"token":     stored.Token,
		"userid":    normalizeUserIDString(stored.UserID),
		"dfid":      stored.DFID,
		"vip_token": stored.VipToken,
		"vip_type":  stored.VipType,
	}
}

// normalizeUserIDString 把科学计数法形式的 userid（如 "8.63006722e+08"）
// 规整为十进制整数字符串；空串或无法解析的值原样返回。
func normalizeUserIDString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return strconv.FormatInt(int64(f), 10)
	}
	return raw
}

// clearSession 清除本地会话，登出时调用。
func clearSession() {
	_ = os.Remove(sessionPath())
}
