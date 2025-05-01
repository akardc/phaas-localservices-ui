package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/dockerclient"
	"phaas-localservices-ui/scheduler"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type apiController struct {
	ctx context.Context

	name                      string
	path                      string
	statusNotificationChannel string
	dir                       os.DirEntry

	jobScheduler *scheduler.Scheduler
	appSettings  *app.Settings

	startedPID   int
	latestStatus Status
}

func (this *apiController) GetBasicDetails() BasicDetails {
	return BasicDetails{
		Name:                      this.name,
		Path:                      this.path,
		StatusNotificationChannel: this.GetStatusNotificationChannel(),
	}
}

func (this *apiController) GetLastModifiedTime() (time.Time, error) {
	dirInfo, err := this.dir.Info()
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "error getting repo details")
		return time.Time{}, fmt.Errorf("error getting repo details: %w", err)
	}
	return dirInfo.ModTime(), nil
}

func (this *apiController) GetActiveBranch() (string, error) {
	repo, err := git.PlainOpen(this.path)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "error opening repo")
		return "", fmt.Errorf("error opening repo: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Failed to get HEAD")
	}
	branch := ""
	if head != nil {
		branch = head.Name().Short()
	}
	return branch, nil
}

func (this *apiController) GetStatus() (Status, error) {

	foundProcess := false
	if this.startedPID != 0 {
		proc, err := os.FindProcess(this.startedPID)
		if err != nil {
			this.startedPID = 0
			if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrProcessDone) {
				slog.With(slog.Any("error", err)).Error("Failed to find started process")
				// not returning so we can fall back to docker lookup
			}
		} else if proc != nil {
			foundProcess = true
		}
	}

	status, err := dockerclient.GetStatus(this.ctx, this.name)
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Error getting docker container status")
		return Status{}, fmt.Errorf("error getting docker container status: %w", err)
	}
	if status == nil || status.State == nil {
		if foundProcess {
			return Status{State: StateStarting}, nil
		}
		return Status{State: StateStopped}, nil
	}
	if status.State.Running {
		return Status{State: StateRunning}, nil
	} else {
		return Status{State: StateStopped}, nil
	}
}

func (this *apiController) GetStatusNotificationChannel() string {
	return fmt.Sprintf("events-%s-status", this.name)
}

func (this *apiController) RegisterStatusWatcher() error {
	jobName := fmt.Sprintf("%s-status-watcher", this.name)
	err := this.jobScheduler.AddJob(jobName, 30*time.Second, func() {
		err := this.refreshStatus()
		if err != nil {
			slog.With(slog.Any("error", err), slog.String("repo", this.name)).Error("Error refreshing status for repo")
		}
	})
	if err != nil && !errors.Is(err, scheduler.ErrJobAlreadyExists) {
		slog.With(slog.Any("error", err)).Error("Error refreshing status for repo")
		return fmt.Errorf("error adding status watcher job: %w", err)
	}
	return nil
}

func (this *apiController) startLowLatencyStatusWatcher() {
	jobName := fmt.Sprintf("%s-status-watcher-low-latency", this.name)
	err := this.jobScheduler.AddJob(jobName, 1*time.Second, func() {
		err := this.refreshStatus()
		if err != nil {
			slog.With(slog.Any("error", err), slog.String("repo", this.name)).Error("Error refreshing status for repo")
		}
		if this.latestStatus.State == StateStopped {
			slog.InfoContext(this.ctx, "Stopping low-latency status watcher")
			this.jobScheduler.RemoveJob(jobName)
		}
	})
	if err != nil && !errors.Is(err, scheduler.ErrJobAlreadyExists) {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "failed to start low latency status watcher")
	}
}

func (this *apiController) refreshStatus() error {

	newStatus, err := this.GetStatus()
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(this.ctx, "Error getting status for repo")
		return fmt.Errorf("error getting status for repo: %w", err)
	}

	statusChanged := false
	if this.latestStatus != newStatus {
		statusChanged = true
	}
	this.latestStatus = newStatus
	if statusChanged {
		runtime.EventsEmit(this.ctx, this.GetStatusNotificationChannel(), this.latestStatus)
	}
	return nil
}

func (this *apiController) Start() error {

	err := this.startMysql()
	if err != nil {
		return err
	}

	status, err := dockerclient.GetStatus(this.ctx, this.name)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		slog.With(slog.Any("error", err)).Error("Error getting repo status")
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status != nil && status.State.Running {
		slog.With(slog.String("repo", this.name)).Info("Already running")
		return nil
	}

	cmd := exec.Command("mage", "serve")
	cmd.Dir = this.path

	repoDataPath := fmt.Sprintf("%s/%s", this.appSettings.DataDirPath, this.name)
	err = os.Mkdir(repoDataPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.With(slog.Any("error", err)).Error("Error creating repo data directory")
		return fmt.Errorf("failed to create data dir: %w", err)
	}
	logFilePath := fmt.Sprintf("%s/service.log", repoDataPath)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		slog.With(slog.Any("error", err)).Error("Error opening repo log file")
		return fmt.Errorf("failed to open log file: %w", err)
	}

	cmd.Env = append(cmd.Environ(), "PHAAS_DOCKER_DISABLE_INTERACTIVE=1")
	overrides := this.appSettings.GetEnvParamOverrides()
	envParams := make([]string, 0, len(overrides))
	for _, param := range overrides {
		if param.Enabled {
			envParams = append(envParams, fmt.Sprintf("PHAAS_OVERRIDE_%s=%s", strings.ToUpper(param.Key), param.Value))
		}
	}
	cmd.Env = append(cmd.Env, envParams...)

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		slog.With(slog.Any("error", err)).Error("Failed to start repo")
		return fmt.Errorf("failed to start repo: %w", err)
	}
	if cmd.Process != nil {
		this.startedPID = cmd.Process.Pid
	}
	err = this.refreshStatus()
	if err != nil {
		slog.With(slog.String("repo", this.name)).Error("Error refreshing status for repo")
	}

	this.startLowLatencyStatusWatcher()

	return nil
}

func (this *apiController) Stop() error {
	status, err := dockerclient.GetStatus(this.ctx, this.name)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		slog.With(slog.Any("error", err)).Error("Error getting repo status")
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status == nil || !status.State.Running {
		slog.With(slog.String("repo", this.name)).Info("Already stopped")
		this.startedPID = 0
		return nil
	}
	err = dockerclient.StopContainer(this.ctx, this.name)
	if err != nil {
		slog.With(slog.Any("error", err), slog.String("repo", this.name)).Info("Failed to stop container")
		return fmt.Errorf("error stopping repo: %w", err)
	}
	this.startedPID = 0
	return nil
}

func (this *apiController) startMysql() error {
	mysqlName := this.name + "-mysql"
	status, err := dockerclient.GetStatus(this.ctx, mysqlName)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		slog.With(slog.Any("error", err)).Error("Error getting repo status")
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status != nil && status.State.Running {
		slog.With(slog.String("repo", this.name)).Info("Mysql already running")
		return nil
	}

	cmd := exec.Command("mage", "mysqlup")
	cmd.Dir = this.path
	cmd.Env = append(cmd.Environ(), "PHAAS_DOCKER_DISABLE_INTERACTIVE=1")

	err = cmd.Run()
	if err != nil {
		slog.With(slog.Any("error", err)).Error("Failed to start mysql")
		return fmt.Errorf("failed to start mysql: %w", err)
	}

	return nil
}
