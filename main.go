package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/repo"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func SetupLogger() logger.Logger {
	dir, err := app.GetSettingsDir()
	if err != nil {
		panic(fmt.Errorf("could not get user config dir: %w", err))
	}
	err = os.Mkdir(dir, os.ModeDir)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(fmt.Errorf("could not create user config dir: %w", err))
	}
	logFile, err := os.Create(filepath.Join(dir, "logs.json"))
	if err != nil {
		panic(fmt.Errorf("could not create log file: %w", err))
	}

	multiwriter := io.MultiWriter(os.Stdout, logFile)
	fileLogger := slog.New(slog.NewJSONHandler(multiwriter, &slog.HandlerOptions{}))
	slog.SetDefault(fileLogger)
	slogLogger := SlogLogger{
		logger: fileLogger,
	}
	return slogLogger
}

type SlogLogger struct {
	logger *slog.Logger
}

func (l SlogLogger) Print(message string) {
	l.logger.Info(message)
}

func (l SlogLogger) Info(message string) {
	l.logger.Info(message)
}

func (l SlogLogger) Error(message string) {
	l.logger.Error(message)
}

func (l SlogLogger) Warning(message string) {
	l.logger.Warn(message)
}

func (l SlogLogger) Panic(message string) {
	l.logger.Error(message)
	panic(message)
}

func (l SlogLogger) Fatal(message string) {
	l.logger.Error(message)
	os.Exit(1)
}

func (l SlogLogger) Debug(message string) {
	l.logger.Debug(message)
}

func (l SlogLogger) Trace(message string) {
	l.logger.Debug(message)
}

func main() {
	// Create an instance of the app structure
	a := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "phaas-localservices-ui",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger:           SetupLogger(),
		LogLevel:         logger.INFO,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        a.startup,
		Bind:             a.getExposedInterfaces(),
		EnumBind: []interface{}{
			repo.AllStates,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
