package repobrowser

import (
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/repo"
	"phaas-localservices-ui/scheduler"
	"slices"
	"strings"
)

type RepoStore struct {
	repoControllers map[string]repo.Controller
}

func (this *RepoStore) Push(name string, controller repo.Controller) {
	if this.repoControllers == nil {
		this.repoControllers = map[string]repo.Controller{}
	}
	this.repoControllers[name] = controller
}

func (this *RepoStore) List() iter.Seq2[string, repo.Controller] {
	return func(yield func(string, repo.Controller) bool) {
		for name, r := range this.repoControllers {
			if !yield(name, r) {
				return
			}
		}
	}
}

var ErrRepoNotFound = fmt.Errorf("repo not found")

func (this *RepoStore) Get(name string) (repo.Controller, error) {
	repoController, found := this.repoControllers[name]
	if !found || repoController == nil {
		return nil, ErrRepoNotFound
	}
	return repoController, nil
}

type RepoBrowser struct {
	ctx      context.Context
	settings *app.Settings

	jobScheduler          *scheduler.Scheduler
	repoControllerFactory *repo.Factory

	repos RepoStore
}

func NewRepoBrowser(
	appSettings *app.Settings,
	jobScheduler *scheduler.Scheduler,
	repoControllerFactory *repo.Factory,
) *RepoBrowser {
	return &RepoBrowser{
		settings:              appSettings,
		jobScheduler:          jobScheduler,
		repos:                 RepoStore{},
		repoControllerFactory: repoControllerFactory,
	}
}

func (this *RepoBrowser) Startup(ctx context.Context) {
	this.ctx = ctx
	err := this.initRepos()
	if err != nil {
		panic(fmt.Errorf("failed to init repos: %w", err))
	}
}

type ListReposOptions struct {
	NameRegex string `json:"nameRegex"`
}

func (this *RepoBrowser) initRepos() error {
	folders, err := os.ReadDir(this.settings.ReposDirPath)
	if err != nil {
		return fmt.Errorf("failed to read repos: %w", err)
	}

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		repoName := folder.Name()
		path := filepath.Join(this.settings.ReposDirPath, repoName)
		repoController := this.repoControllerFactory.BuildRepoController(this.ctx, path, repoName, folder)
		if repoController != nil {
			this.repos.Push(repoName, repoController)
		}
	}
	return nil
}

type Filter struct{}

func (this *RepoBrowser) ListRepos() ([]repo.BasicDetails, error) {
	list := make([]repo.BasicDetails, 0)
	for _, repoController := range this.repos.List() {
		list = append(list, repoController.GetBasicDetails())
	}
	slices.SortFunc(list, func(a, b repo.BasicDetails) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return list, nil
}

func (this *RepoBrowser) GetRepoStatus(repoName string) (repo.Status, error) {
	repoController, err := this.repos.Get(repoName)
	if err != nil {
		return repo.Status{}, fmt.Errorf("failed to get repo '%s': %w", repoName, err)
	}
	status, err := repoController.GetStatus()
	if err != nil {
		return repo.Status{}, fmt.Errorf("failed to get repo '%s': %w", repoName, err)
	}
	return status, nil
}

func (this *RepoBrowser) StartRepo(repoName string) error {
	repoController, err := this.repos.Get(repoName)
	if err != nil {
		return fmt.Errorf("failed to get repo '%s': %w", repoName, err)
	}
	err = repoController.Start()
	if err != nil {
		return fmt.Errorf("failed to start repo '%s': %w", repoName, err)
	}
	return nil
}

func (this *RepoBrowser) StopRepo(repoName string) error {
	repoController, err := this.repos.Get(repoName)
	if err != nil {
		return fmt.Errorf("failed to get repo '%s': %w", repoName, err)
	}
	err = repoController.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop repo '%s': %w", repoName, err)
	}
	return nil
}

func (this *RepoBrowser) GetRepoRepoStatusNotificationChannel(repoName string) (string, error) {
	repoController, err := this.repos.Get(repoName)
	if err != nil {
		return "", fmt.Errorf("failed to get repo '%s': %w", repoName, err)
	}
	return repoController.GetStatusNotificationChannel(), nil
}

func (this *RepoBrowser) RegisterRepoStatusWatcher(repoName string) error {
	repoController, err := this.repos.Get(repoName)
	if err != nil {
		return fmt.Errorf("failed to get repo '%s': %w", repoName, err)
	}
	return repoController.RegisterStatusWatcher()
}
