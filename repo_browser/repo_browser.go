package repobrowser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"phaas-localservices-ui/repo"
	"phaas-localservices-ui/settings"
	"regexp"
	"slices"
	"time"
)

type RepoBrowser struct {
	ctx      context.Context
	settings *settings.Settings

	repoMap map[string]*repo.Repo
	repos   []*repo.Repo
}

func NewRepoBrowser(
	settings *settings.Settings,
) *RepoBrowser {
	return &RepoBrowser{
		settings: settings,
		repoMap:  make(map[string]*repo.Repo),
	}
}

func (this *RepoBrowser) Startup(ctx context.Context) {
	this.ctx = ctx
	err := this.initRepos()
	if err != nil {
		panic(fmt.Errorf("failed to init repos: %w", err))
	}
}

type RepoInfo struct {
	Name         string    `json:"name"`
	LastModified time.Time `json:"lastModified"`
	Branch       string    `json:"branch"`
}

type ListReposOptions struct {
	NameRegex string `json:"nameRegex"`
}

func (this *RepoBrowser) initRepos() error {
	folders, err := os.ReadDir(this.settings.ReposDirPath)
	if err != nil {
		return fmt.Errorf("failed to read repos: %w", err)
	}

	nameRegex := regexp.MustCompile("phaas-.*-((ui)|(api))")
	for _, folder := range folders {
		if !folder.IsDir() || !nameRegex.MatchString(folder.Name()) {
			continue
		}
		path := filepath.Join(this.settings.ReposDirPath, folder.Name())
		r, err := repo.NewRepo(folder, path, this.settings)
		if err != nil {
			return fmt.Errorf("failed to build repo: %w", err)
		}
		this.repos = append(this.repos, r)
		this.repoMap[folder.Name()] = r
	}
	return nil
}

func (this *RepoBrowser) ListRepos(opts *ListReposOptions) []RepoInfo {
	repos := make([]RepoInfo, 0, len(this.repos))
	for _, r := range this.repos {
		info := r.BasicInfo()
		repos = append(repos, RepoInfo{
			Name:         info.Name,
			LastModified: info.LastModified,
			Branch:       info.Branch,
		})
	}
	slices.SortFunc(repos, func(a, b RepoInfo) int {
		// sort by lastModified descending
		return b.LastModified.Compare(a.LastModified)
	})
	return repos
}

type RepoStatus struct {
	BranchName string `json:"branchName"`
}

// func (this *RepoBrowser) RepoStatus(path string) RepoStatus {
// 	ctx := slogctx.Append(this.ctx, slog.String("path", path))
// 	repo, err := git.PlainOpen(path)
// 	if err != nil {
// 		return RepoStatus{}
// 	}
// 	head, err := repo.Head()
// 	if err != nil {
// 		slog.With(slog.Any("error", err)).ErrorContext(ctx, "Failed to get HEAD")
// 		return RepoStatus{}
// 	}
// 	return RepoStatus{
// 		BranchName: head.Name().Short(),
// 	}
// }

func (this *RepoBrowser) StartRepo(name string) error {
	repo := this.repoMap[name]
	if repo == nil {
		return fmt.Errorf("repo not found: %s", name)
	}

	err := repo.Start()
	if err != nil {
		return fmt.Errorf("failed to start repo %s: %w", name, err)
	}
	return nil
}

func (this *RepoBrowser) StopRepo(name string) error {
	repo := this.repoMap[name]
	if repo == nil {
		return fmt.Errorf("repo not found: %s", name)
	}

	err := repo.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop repo %s: %w", name, err)
	}
	return nil
}
