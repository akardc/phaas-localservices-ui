package main

import (
	"context"
	"log/slog"
	"os"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/mage"
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
	appSettings := &app.Settings{}
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

	err := a.appSettings.Startup(ctx)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(ctx, "failed to load app settings")
		os.Exit(1)
	}
	err = mage.Init(ctx, a.appSettings)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(ctx, "Failed to init shell")
		os.Exit(1)
	}
	a.repoBrowser.Startup(ctx)
	a.jobScheduler.Start(ctx)
}

func (a *App) getExposedInterfaces() []any {
	return []any{
		a.repoBrowser,
		a.appSettings,
	}
}
