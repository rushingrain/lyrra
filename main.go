package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"kuwave/internal/auth"
	"kuwave/internal/banner"
	"kuwave/internal/log"
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
		},
	})

	if err != nil {
		logger.Error("app startup failed", "err", err.Error())
	}
}
