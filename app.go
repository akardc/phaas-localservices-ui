package main

import (
	"context"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/repo"
	repobrowser "phaas-localservices-ui/repo_browser"
	"phaas-localservices-ui/scheduler"
)

// App struct
type App struct {
	ctx context.Context

	jobScheduler *scheduler.Scheduler
	appSettings  *app.Settings
	repoFactory  *repo.Factory
	repoBrowser  *repobrowser.RepoBrowser
}

// NewApp creates a new App application struct
func NewApp() *App {
	jobScheduler := scheduler.New()
	appSettings := &app.Settings{
		ReposDirPath: "/Users/cakard/go/src/github.com/BidPal",
		DataDirPath:  "/users/cakard/Documents/phaas-localservices-ui",
	}
	repoFactory := repo.NewFactory(appSettings, jobScheduler)

	return &App{
		jobScheduler: jobScheduler,
		appSettings:  appSettings,
		repoFactory:  repoFactory,
		repoBrowser:  repobrowser.NewRepoBrowser(appSettings, jobScheduler, repoFactory),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.repoBrowser.Startup(ctx)
}

func (a *App) getExposedInterfaces() []any {
	return []any{a.repoBrowser}
}
