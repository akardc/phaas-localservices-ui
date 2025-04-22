package repo_bak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"phaas-localservices-ui/app"
	"time"

	"github.com/go-git/go-git/v5"
	slogctx "github.com/veqryn/slog-context"
)

type Repo struct {
	id           string
	name         string
	path         string
	lastModified time.Time
	dir          os.DirEntry

	logFilePath  string
	dataFilePath string

	data *repoData

	settings *app.Settings
}

type repoData struct {
	RunPID int `json:"pid"`
}

func NewRepo(dir os.DirEntry, path string, settings *app.Settings) (*Repo, error) {

	dataDirPath := fmt.Sprintf("%s/%s", settings.DataDirPath, dir.Name())
	if err := os.Mkdir(dataDirPath, os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("unable to create data directory: %w", err)
	}
	logFilePath := fmt.Sprintf("%s/service.log", dataDirPath)
	dataFilePath := fmt.Sprintf("%s/data.json", dataDirPath)

	repoData := repoData{}

	dataFile, err := os.ReadFile(dataFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	err = json.Unmarshal(dataFile, &repoData)
	if err != nil {
		slog.With(slog.Any("error", err)).Error("failed to parse config file - assuming first time connecting to repo")
	}
	if repoData.RunPID != 0 {
		slog.With(slog.Int("pid", repoData.RunPID)).Info("Process PID found for repo - will attempt to connect")
		_, err := os.FindProcess(repoData.RunPID)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			slog.With(slog.Int("pid", repoData.RunPID)).Info("Process PID not found - it was likely killed")
			repoData.RunPID = 0
		} else if err != nil {
			slog.With(slog.Any("error", err)).Error("Failed to find existing process for repo")
		}
	}

	dataJSON, err := json.Marshal(repoData)
	if err != nil {
		slog.With(slog.Any("error", err)).Error("failed to marshal repo data")
	} else {
		err = os.WriteFile(dataFilePath, dataJSON, os.ModePerm)
		if err != nil {
			slog.With(slog.Any("error", err)).Error("failed to store repo data")
		}
	}

	return &Repo{
		name:         dir.Name(),
		path:         path,
		dir:          dir,
		logFilePath:  logFilePath,
		dataFilePath: dataFilePath,
		data:         &repoData,
		settings:     settings,
	}, nil
}

type BasicInfo struct {
	Name         string    `json:"name"`
	LastModified time.Time `json:"lastModified"`
	Branch       string    `json:"branch"`
}

func (this *Repo) BasicInfo() BasicInfo {
	return BasicInfo{
		Name:         this.name,
		LastModified: this.lastModified,
		Branch:       this.getBranch(),
	}
}

func (this *Repo) getBranch() string {
	ctx := slogctx.Append(context.Background(), slog.String("path", this.path))
	repo, err := git.PlainOpen(this.path)
	if err != nil {
		return ""
	}
	head, err := repo.Head()
	if err != nil {
		slog.With(slog.Any("error", err)).ErrorContext(ctx, "Failed to get HEAD")
		return ""
	}
	return head.Name().Short()
}

type RunningStatus string

const (
	Running RunningStatus = "Running"
	Stopped RunningStatus = "Stopped"
	Unknown RunningStatus = "Unknown"
)

var AllRunningStatus = []struct {
	Value  RunningStatus
	TSName string
}{
	{Running, "Running"},
	{Stopped, "Stopped"},
	{Unknown, "Unknown"},
}

func (this *Repo) CheckStatus() RunningStatus {
	if this.data == nil {
		dataFile, err := os.ReadFile(this.dataFilePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.With(slog.Any("error", err)).Error("failed to open data file")
			return Unknown
		}
		err = json.Unmarshal(dataFile, this.data)
		if err != nil {
			slog.With(slog.Any("error", err)).Error("failed to parse data file")
			return Unknown
		}
	}

	defer func() {
		dataJSON, err := json.Marshal(this.data)
		if err != nil {
			slog.With(slog.Any("error", err)).Error("failed to marshal repo data")
			return
		}
		err = os.WriteFile(this.dataFilePath, dataJSON, os.ModePerm)
		if err != nil {
			slog.With(slog.Any("error", err)).Error("failed to store repo data")
		}
	}()

	if this.data.RunPID != 0 {
		slog.With(slog.Int("pid", this.data.RunPID)).Info("Process PID found for repo - will attempt to connect")
		_, err := os.FindProcess(this.data.RunPID)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			slog.With(slog.Int("pid", this.data.RunPID)).Info("Process PID not found - it was likely killed")
			this.data.RunPID = 0
		} else if err != nil {
			slog.With(slog.Any("error", err)).Error("Failed to find existing process for repo")
			return Unknown
		}
	}

	if this.data.RunPID == 0 {
		return Stopped
	} else if this.data.RunPID > 0 {
		return Running
	}
	return Unknown
}

func (this *Repo) Start() error {
	if this.data != nil && this.data.RunPID != 0 {
		return errors.New("repo already started")
	}

	cmd := exec.Command("mage", "serve")
	cmd.Dir = this.path

	err := os.Mkdir(this.settings.DataDirPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create data dir: %w", err)
	}
	logFile, err := os.Create(this.logFilePath)
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

	if this.data == nil {
		this.data = &repoData{}
	}
	this.data.RunPID = cmd.Process.Pid
	return nil
}

func (this *Repo) Stop() error {
	if this.data == nil || this.data.RunPID == 0 {
		slog.Info("No process to stop")
		return nil
	}

	proc, err := os.FindProcess(this.data.RunPID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if proc != nil {
		err := proc.Kill()
		if err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	return nil
}
