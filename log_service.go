package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Service 统一日志服务。
//
// 后端使用标准库 log/slog 输出结构化日志，写入用户配置目录下的
// logs/kuwave.log；由 lumberjack 按大小轮转并自动清理，磁盘占用有界。
// 同时提供 Wails logger.Logger 适配器，将 Wails 运行时内部日志路由到同一文件。
type Service struct {
	logger *slog.Logger
	writer io.WriteCloser
}

// New 初始化日志服务。
//   - 日志目录：os.UserConfigDir()/kuwave/logs（与 session.go 的会话目录同基址）
//   - 写入器：lumberjack，单文件上限 10MB、保留 3 个备份、最长 7 天、过期压缩
//   - 级别：默认 INFO，过滤 DEBUG 噪音
func New() *Service {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	logDir := filepath.Join(dir, "kuwave", "logs")
	_ = os.MkdirAll(logDir, 0o700)
	return newService(filepath.Join(logDir, "kuwave.log"))
}

// newService 用指定文件构造日志服务，便于测试注入临时目录。
func newService(filename string) *Service {
	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     7, // 天
		Compress:   true,
	}

	return &Service{
		logger: slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})),
		writer: writer,
	}
}

// Logger 返回底层 slog.Logger，供 App / AuthService 等注入使用。
func (s *Service) Logger() *slog.Logger {
	return s.logger
}

// Close 关闭日志写入器，应用退出前调用。
func (s *Service) Close() error {
	return s.writer.Close()
}

// wailsAdapter 实现 wails logger.Logger 接口，将 Wails 运行时内部日志路由到 slog。
type wailsAdapter struct {
	logger *slog.Logger
}

// NewWailsAdapter 构造 Wails 日志适配器，用于 options.Logger。
func NewWailsAdapter(service *Service) logger.Logger {
	return &wailsAdapter{logger: service.logger}
}

func (a *wailsAdapter) Print(message string)   { a.logger.Info(message) }
func (a *wailsAdapter) Trace(message string)   { a.logger.Debug(message) }
func (a *wailsAdapter) Debug(message string)   { a.logger.Debug(message) }
func (a *wailsAdapter) Info(message string)    { a.logger.Info(message) }
func (a *wailsAdapter) Warning(message string) { a.logger.Warn(message) }
func (a *wailsAdapter) Error(message string)   { a.logger.Error(message) }

// Fatal 降级为 Error：日志不应导致桌面应用进程退出。
func (a *wailsAdapter) Fatal(message string) { a.logger.Error(message) }
