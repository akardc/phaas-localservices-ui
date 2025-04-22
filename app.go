package main

import (
	"context"
	"fmt"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/repo"
	repobrowser "phaas-localservices-ui/repo_browser"
)

// App struct
type App struct {
	ctx context.Context

	appSettings *app.Settings
	repoFactory *repo.Factory
	repoBrowser *repobrowser.RepoBrowser
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
