package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 日志最先初始化：slog 写入 logs/kuwave.log（lumberjack 轮转清理），
	// 并注入 App / AuthService，供关键业务事件记录。
	logService := New()
	log := logService.Logger()
	defer func() {
		if err := logService.Close(); err != nil {
			log.Error("failed to close log service", "err", err.Error())
		}
	}()

	// Create an instance of the app structure
	app := NewApp(log)

	// Create the auth service instance
	authService, err := NewAuthService(log)
	if err != nil {
		log.Error("failed to init auth service", "err", err.Error())
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
		Logger:           NewWailsAdapter(logService),
		Bind: []interface{}{
			app,
			authService,
		},
	})

	if err != nil {
		log.Error("app startup failed", "err", err.Error())
	}
}
