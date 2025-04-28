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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type repoDetails struct {
	name                      string
	path                      string
	statusUpdatedEventChannel string
	dir                       os.DirEntry
}

type apiController struct {
	ctx context.Context

	repoDetails repoDetails

	jobScheduler *scheduler.Scheduler
	appSettings  *app.Settings

	latestStatus Status
}

func (this *apiController) GetBasicDetails() BasicDetails {
	return BasicDetails{
		Name:                      this.repoDetails.name,
		Path:                      this.repoDetails.path,
		StatusUpdatedEventChannel: this.repoDetails.statusUpdatedEventChannel,
	}
}

func (this *apiController) GetLastModifiedTime() (time.Time, error) {
	dirInfo, err := this.repoDetails.dir.Info()
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting repo details: %w", err)
	}
	return dirInfo.ModTime(), nil
}

func (this *apiController) GetActiveBranch() (string, error) {
	repo, err := git.PlainOpen(this.repoDetails.path)
	if err != nil {
		return "", fmt.Errorf("error opening repo: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		slog.With(slog.Any("error", err)).Error("Failed to get HEAD")
	}
	branch := ""
	if head != nil {
		branch = head.Name().Short()
	}
	return branch, nil
}

func (this *apiController) GetStatus() (Status, error) {
	status, err := dockerclient.GetStatus(this.ctx, this.repoDetails.name)
	if err != nil {
		return Status{}, fmt.Errorf("error getting docker container status: %w", err)
	}
	if status.State == nil {
		return Status{State: StateUnknown}, nil
	}
	if status.State.Running {
		return Status{State: StateRunning}, nil
	} else {
		return Status{State: StateStopped}, nil
	}
}

func (this *apiController) GetStatusNotificationChannel() string {
	return fmt.Sprintf("events-%s-status", this.repoDetails.name)
}

func (this *apiController) RegisterStatusWatcher() error {
	jobName := fmt.Sprintf("%s-status-watcher", this.repoDetails.name)
	err := this.jobScheduler.AddJob(jobName, 30*time.Second, func() {
		err := this.refreshStatus()
		if err != nil {
			slog.With(slog.Any("error", err), slog.String("repo", this.repoDetails.name)).Error("Error refreshing status for repo")
		}
	})
	if err != nil && !errors.Is(err, scheduler.ErrJobAlreadyExists) {
		return fmt.Errorf("error adding status watcher job: %w", err)
	}
	return nil
}

func (this *apiController) refreshStatus() error {

	newStatus, err := this.GetStatus()
	if err != nil {
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

	status, err := dockerclient.GetStatus(this.ctx, this.repoDetails.name)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status != nil && status.State.Running {
		slog.With(slog.String("repo", this.repoDetails.name)).Info("Already running")
		return nil
	}

	cmd := exec.Command("mage", "serve")
	cmd.Dir = this.repoDetails.path

	repoDataPath := fmt.Sprintf("%s/%s", this.appSettings.DataDirPath, this.repoDetails.name)
	err = os.Mkdir(repoDataPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create data dir: %w", err)
	}
	logFilePath := fmt.Sprintf("%s/service.log", repoDataPath)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// TODO: need to add support for this in mage
	cmd.Env = append(cmd.Environ(), "PHAAS_DOCKER_DISABLE_INTERACTIVE=1")

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start repo: %w", err)
	}

	return nil
}

func (this *apiController) Stop() error {
	status, err := dockerclient.GetStatus(this.ctx, this.repoDetails.name)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status == nil || !status.State.Running {
		slog.With(slog.String("repo", this.repoDetails.name)).Info("Already stopped")
		return nil
	}
	err = dockerclient.StopContainer(this.ctx, this.repoDetails.name)
	if err != nil {
		slog.With(slog.Any("error", err), slog.String("repo", this.repoDetails.name)).Info("Failed to stop container")
		return fmt.Errorf("error stopping repo: %w", err)
	}
	return nil
}

func (this *apiController) startMysql() error {
	mysqlName := this.repoDetails.name + "-mysql"
	status, err := dockerclient.GetStatus(this.ctx, mysqlName)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status != nil && status.State.Running {
		slog.With(slog.String("repo", this.repoDetails.name)).Info("Mysql already running")
		return nil
	}

	cmd := exec.Command("mage", "mysqlup")
	cmd.Dir = this.repoDetails.path
	// TODO: need to add support for this in mage
	cmd.Env = append(cmd.Environ(), "PHAAS_DOCKER_DISABLE_INTERACTIVE=1")

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start mysql: %w", err)
	}

	return nil
}
