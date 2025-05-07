package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const ReposLocationChangedEvent = "repos-location-changed"

type Settings struct {
	ctx context.Context

	ReposDirPath        string `json:"reposDirPath"`
	DataDirPath         string `json:"dataDirPath"`
	ShellExecutablePath string `json:"shellExecutablePath"`
	ShellInitFilePath   string `json:"shellInitFilePath"`

	EnvParams []EnvParam `json:"envParams"`

	settingsPath string
}

func GetSettingsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		slog.With(slog.Any("error", err)).Error("Could not get user config dir")
		return "", fmt.Errorf("could not get user config dir: %w", err)
	}
	return filepath.Join(configDir, "phaas-localservices-manager"), nil
}

func (this *Settings) Startup(ctx context.Context) error {
	this.ctx = ctx

	settingsDir, err := GetSettingsDir()
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Could not get user settings dir")
		return fmt.Errorf("could not get user settings dir: %w", err)
	}
	this.settingsPath = filepath.Join(settingsDir, "settings.json")
	if _, err := os.Stat(this.settingsPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(settingsDir, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Could not create settings directory")
			return fmt.Errorf("could not create settings directory: %w", err)
		}
		_, err = os.Create(this.settingsPath)
		if err != nil && !errors.Is(err, os.ErrExist) {
			slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Could not create settings file")
			return fmt.Errorf("could not create settings file: %w", err)
		}
	}

	settingsJSON, err := os.ReadFile(this.settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Failed to read settings")
		return fmt.Errorf("failed to read settings: %w", err)
	}

	if len(settingsJSON) == 0 {
		return nil
	}

	err = json.Unmarshal(settingsJSON, this)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Failed to unmarshal settings.json")
		return fmt.Errorf("failed to unmarshal settings.json: %w", err)
	}
	slog.With(slog.Any("settings", this)).InfoContext(ctx, "Loaded with settings")

	return nil
}

func (this *Settings) GetSettings() Settings {
	return *this
}

func (this *Settings) SaveSettings(settings Settings) error {
	reposDirChanged := this.ReposDirPath != settings.ReposDirPath
	this.ReposDirPath = settings.ReposDirPath
	this.DataDirPath = settings.DataDirPath
	this.EnvParams = settings.EnvParams

	err := this.writeToFile()
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Failed to save app settings")
		return fmt.Errorf("failed to save app settings: %w", err)
	}
	if reposDirChanged {
		runtime.EventsEmit(this.ctx, ReposLocationChangedEvent)
	}
	return nil
}

type EnvParam struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

func (this *Settings) GetEnvParamOverrides() []EnvParam {
	return this.EnvParams
}

func (this *Settings) writeToFile() error {
	settingsJSON, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Failed to marshal settings to json to save to file")
		return fmt.Errorf("failed to marshal settings to json to save to file: %s", err)
	}

	err = os.WriteFile(this.settingsPath, settingsJSON, os.ModePerm)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Failed to save settings to file")
		return fmt.Errorf("failed to save settings to file: %s", err)
	}

	return nil
}
