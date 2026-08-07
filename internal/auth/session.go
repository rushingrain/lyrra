package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 会话持久化：只保存关键的登录态 cookie 字段。
// 存储位置为用户配置目录（macOS: ~/Library/Application Support/kuwave/、
// Windows: %AppData%\kuwave\、Linux: ~/.config/kuwave/），
// 不依赖当前工作目录（桌面应用 cwd 不可靠）。

func sessionPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "kuwave", "session.json")
}

// saveSession 保存会话，登录成功后调用。
// 固定写入 token / userid / vip_token / vip_type 四个登录态 cookie 字段。
// 非 VIP 账号的 vip_token / vip_type 可能为空，但保持一致的结构便于后续扩展。
func saveSession(cookie map[string]string) {
	essential := map[string]string{
		"token":     cookie["token"],
		"userid":    normalizeUserIDString(cookie["userid"]),
		"vip_token": cookie["vip_token"],
		"vip_type":  cookie["vip_type"],
	}
	data, err := json.Marshal(essential)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(sessionPath()), 0o700); err != nil {
		return
	}
	_ = os.WriteFile(sessionPath(), data, 0o600)
}

// loadSession 恢复会话，初始化 client 时调用。无会话时返回 nil。
func loadSession() map[string]string {
	data, err := os.ReadFile(sessionPath())
	if err != nil {
		return nil
	}
	var cookie map[string]string
	if err := json.Unmarshal(data, &cookie); err != nil {
		return nil
	}
	// 兼容历史会话：早期扫码登录可能把 userid 存成科学计数法
	// （如 "8.63006722e+08"），加载时统一规整为十进制整数。
	cookie["userid"] = normalizeUserIDString(cookie["userid"])
	return cookie
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
