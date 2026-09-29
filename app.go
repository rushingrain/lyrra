package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	log *slog.Logger
	// mu 保护 ctx：startup 在 Wails 启动时写入，单实例回调在另一个 goroutine 读取。
	mu  sync.Mutex
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp(log *slog.Logger) *App {
	return &App{log: log}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()
	a.log.Info("app started")
}

// focusWindow 把已有窗口从最小化状态恢复并置前。
// 第二个实例启动时由 main.go 的单实例锁回调调用（未导出，不暴露给前端）。
func (a *App) focusWindow() {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil {
		return
	}
	runtime.WindowUnminimise(ctx)
	runtime.WindowShow(ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
