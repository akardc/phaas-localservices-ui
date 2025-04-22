package repo

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/dockerclient"

	"github.com/go-git/go-git/v5"
)

type repoDetails struct {
	name string
	path string
	dir  os.DirEntry
}

type apiController struct {
	appSettings *app.Settings
	repoDetails repoDetails
}

/*
	TODO
	- running
	- logs
	- get run args
	- set run args

	- start
	- stop
	-
*/

func (this *apiController) GetBasicDetails() BasicDetails {
	return BasicDetails{
		Name: this.repoDetails.name,
		Path: this.repoDetails.path,
	}
}
func (this *apiController) GetStatus() (*Status, error) {

	dirInfo, err := this.repoDetails.dir.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting repo details: %w", err)
	}

	repo, err := git.PlainOpen(this.repoDetails.path)
	if err != nil {
		return nil, fmt.Errorf("error opening repo: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		slog.With(slog.Any("error", err)).Error("Failed to get HEAD")
	}
	branch := ""
	if head != nil {
		branch = head.Name().Short()
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting repo worktree: %w", err)
	}
	repoStatus, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("error getting repo status: %w", err)
	}

	status, err := dockerclient.GetStatus(this.repoDetails.name)
	if err != nil {
		return nil, fmt.Errorf("error getting repo status: %w", err)
	}

	return &Status{
		LastModified: dirInfo.ModTime(),
		Branch:       branch,
		IsClean:      repoStatus.IsClean(),
		Running:      status.Running,
	}, nil
}

func (this *apiController) Start() error {
	status, err := dockerclient.GetStatus(this.repoDetails.name)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status != nil && status.Running {
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

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start repo: %w", err)
	}

	return nil
}

func (this *apiController) Stop() error {
	status, err := dockerclient.GetStatus(this.repoDetails.name)
	if err != nil && !errors.Is(err, dockerclient.ErrNoContainerFound) {
		return fmt.Errorf("error getting repo status: %w", err)
	}
	if status == nil || !status.Running {
		slog.With(slog.String("repo", this.repoDetails.name)).Info("Already stopped")
		return nil
	}
	err = dockerclient.StopContainer(this.repoDetails.name)
	if err != nil {
		slog.With(slog.Any("error", err), slog.String("repo", this.repoDetails.name)).Info("Failed to stop container")
		return fmt.Errorf("error stopping repo: %w", err)
	}
	return nil
}
