package main

import (
	"embed"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"lyrra/internal/auth"
	"lyrra/internal/banner"
	"lyrra/internal/log"
	"lyrra/internal/netguard"
	"lyrra/internal/play"
	"lyrra/internal/playlist"
	"lyrra/internal/search"
	"lyrra/internal/song"
	"lyrra/internal/vip"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 日志最先初始化：slog 写入 logs/lyrra.log（lumberjack 轮转清理），
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

	// Create the daily vip service instance
	vipService, err := vip.New(logger)
	if err != nil {
		logger.Error("failed to init vip service", "err", err.Error())
		return
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "lyrra",
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
			vipService,
		},
		// 单实例锁：桌面上多开时，第二个实例把启动参数交给第一个实例后自行退出。
		// 它是 vip 风控的一部分：vip.json 的“一天只成功一次”依赖同一时刻只有一个
		// 进程读写本地状态（vip 包里的互斥锁只能管住进程内并发）。
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "lyrra",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				logger.Info("second instance launched, focusing existing window",
					"args", strings.Join(data.Args, " "), "workdir", data.WorkingDirectory)
				app.focusWindow()
			},
		},
		Mac: &mac.Options{
			// 隐藏标题文字，保留原生交通灯按钮，内容填满窗口
			TitleBar: mac.TitleBarHiddenInset(),
		},
	})

	if err != nil {
		logger.Error("app startup failed", "err", err.Error())
	}
}
