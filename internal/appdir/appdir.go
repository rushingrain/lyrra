// Package appdir 统一应用名与用户数据目录。
//
// 登录会话、播放队列、vip 记录、日志都落在同一个基目录下，集中在这里定义，
// 避免各包各自拼路径导致改名时漏改。
package appdir

import (
	"os"
	"path/filepath"
)

// appName 同时作为用户数据目录名（macOS: ~/Library/Application Support/lyrra/）。
const appName = "lyrra"

// Dir 返回应用数据目录；取不到用户配置目录时回退当前目录。
func Dir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, appName)
}
