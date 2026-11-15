package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"lyrra/internal/appdir"
	"lyrra/internal/atomicfile"
)

// 默认主题色与前端首次绘制的 CSS 保持一致。
const defaultThemeColor = "#10B981"

var themeColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type themeSettings struct {
	AccentColor string `json:"accentColor"`
}

// GetThemeColor 读取设备级主题色，缺失或损坏的配置回退到默认色。
func (a *App) GetThemeColor() (string, error) {
	a.themeMu.Lock()
	defer a.themeMu.Unlock()

	payload, err := os.ReadFile(filepath.Join(appdir.Dir(), "theme.json"))
	if os.IsNotExist(err) {
		return defaultThemeColor, nil
	}
	if err != nil {
		return "", fmt.Errorf("读取主题设置失败: %w", err)
	}
	var settings themeSettings
	if json.Unmarshal(payload, &settings) != nil || !themeColorPattern.MatchString(settings.AccentColor) {
		return defaultThemeColor, nil
	}
	return strings.ToUpper(settings.AccentColor), nil
}

// SaveThemeColor 校验后原子写入；主题独立于账号，登出时保留。
func (a *App) SaveThemeColor(color string) error {
	color = strings.ToUpper(strings.TrimSpace(color))
	if !themeColorPattern.MatchString(color) {
		return fmt.Errorf("主题色必须为六位 HEX 值，如 #3D8E86")
	}
	payload, err := json.Marshal(themeSettings{AccentColor: color})
	if err != nil {
		return err
	}
	a.themeMu.Lock()
	defer a.themeMu.Unlock()
	if err := atomicfile.Write(filepath.Join(appdir.Dir(), "theme.json"), payload); err != nil {
		return fmt.Errorf("保存主题设置失败: %w", err)
	}
	return nil
}
