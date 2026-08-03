package main

import (
	"context"
	"fmt"
	"log/slog"
)

// App struct
type App struct {
	ctx context.Context
	log *slog.Logger
}

// NewApp creates a new App application struct
func NewApp(log *slog.Logger) *App {
	return &App{log: log}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.log.Info("app started")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
