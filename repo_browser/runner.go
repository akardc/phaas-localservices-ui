package repobrowser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

type Runner struct {
	repoName       string
	path           string
	processPID     int
	logFilePath    string
	configFilePath string

	config config
}

type config struct {
	ProcessPID int `json:"process_pid"`
}

func NewRunner(
	repoName string,
	path string,
) *Runner {
	return &Runner{
		repoName:       repoName,
		path:           path,
		logFilePath:    fmt.Sprintf("/users/cakard/Documents/localservices/%s/logs.txt", repoName),
		configFilePath: fmt.Sprintf("/users/cakard/Documents/localservices/%s/config.json", repoName),
	}
}

func (this *Runner) Start(ctx context.Context) error {
	cmd := exec.Command("mage", "run")
	cmd.Dir = this.path

	configFile, err := os.ReadFile(this.configFilePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	err = json.Unmarshal(configFile, &this.config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	logs, err := os.Create(this.logFilePath)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(ctx, "failed to create log file")
	}
	if logs != nil {
		defer logs.Close()
		cmd.Stdout = logs
		cmd.Stderr = logs
	}
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start repo: %w", err)
	}
	if cmd.Process != nil {
		this.processPID = cmd.Process.Pid
	}
	this.config.ProcessPID = this.processPID
	configJSON, err := json.Marshal(this.config)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(ctx, "failed to marshal config json")
	} else {
		err = os.WriteFile(this.configFilePath, configJSON, os.ModePerm)
		if err != nil {
			slog.With(slog.Any("error", err)).ErrorContext(ctx, "failed to write config file")
		}
	}
	return nil
}
