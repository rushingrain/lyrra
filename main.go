package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"kuwave/internal/auth"
	"kuwave/internal/banner"
	"kuwave/internal/log"
	"kuwave/internal/netguard"
	"kuwave/internal/play"
	"kuwave/internal/playlist"
	"kuwave/internal/search"
	"kuwave/internal/song"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 日志最先初始化：slog 写入 logs/kuwave.log（lumberjack 轮转清理），
	// 并注入 App / auth.Service，供关键业务事件记录。
	logService := log.New()
	logger := logService.Logger()
	defer func() {
		if err := logService.Close(); err != nil {
			logger.Error("failed to close log service", "err", err.Error())
		}
	}()

	// 明文网络护栏：拒绝所有非 loopback 的 http 请求，避免 SDK 内部（例如登录态校验
	// 触发的 token 刷新）把账号凭证发到 http 上。只影响 Go 侧 HTTP 客户端，
	// WebView 加载的音频流不受影响（那部分由 build/darwin/*.plist 的 ATS 白名单控制）。
	netguard.Install(logger)

	// Create an instance of the app structure
	app := NewApp(logger)

	// Create the auth service instance
	authService, err := auth.New(logger)
	if err != nil {
		logger.Error("failed to init auth service", "err", err.Error())
		return
	}

	// Create the banner service instance
	bannerService, err := banner.New(logger)
	if err != nil {
		logger.Error("failed to init banner service", "err", err.Error())
		return
	}

	// Create the playlist service instance
	playlistService, err := playlist.New(logger)
	if err != nil {
		logger.Error("failed to init playlist service", "err", err.Error())
		return
	}

	// Create the song service instance
	songService, err := song.New(logger)
	if err != nil {
		logger.Error("failed to init song service", "err", err.Error())
		return
	}

	// Create the search service instance
	searchService, err := search.New(logger)
	if err != nil {
		logger.Error("failed to init search service", "err", err.Error())
		return
	}

	// Create the play service instance
	playService, err := play.New(logger)
	if err != nil {
		logger.Error("failed to init play service", "err", err.Error())
		return
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "kuwave",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Logger:           log.NewWailsAdapter(logService),
		Bind: []interface{}{
			app,
			authService,
			bannerService,
			playlistService,
			songService,
			searchService,
			playService,
		},
	})

	if err != nil {
		logger.Error("app startup failed", "err", err.Error())
	}
}
