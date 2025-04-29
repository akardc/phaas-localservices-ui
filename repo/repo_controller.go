package repo

import (
	"context"
	"os"
	"phaas-localservices-ui/app"
	"phaas-localservices-ui/scheduler"
	"regexp"
	"time"
)

type Controller interface {
	GetBasicDetails() BasicDetails
	GetLastModifiedTime() (time.Time, error)
	GetActiveBranch() (string, error)
	GetStatus() (Status, error)
	GetStatusNotificationChannel() string
	RegisterStatusWatcher() error
	Start() error
	Stop() error
}

type State string

const (
	StateUnknown  State = "unknown"
	StateStarting State = "starting"
	StateRunning  State = "running"
	StateStopped  State = "stopped"
)

var AllStates = []struct {
	Value  State
	TSName string
}{
	{StateUnknown, "Unknown"},
	{StateStarting, "starting"},
	{StateRunning, "running"},
	{StateStopped, "stopped"},
}

type Status struct {
	State State `json:"state"`
}

type BasicDetails struct {
	Name                      string `json:"name"`
	Path                      string `json:"path"`
	StatusNotificationChannel string `json:"statusNotificationChannel"`
}

type Factory struct {
	settings     *app.Settings
	jobScheduler *scheduler.Scheduler
}

func NewFactory(
	settings *app.Settings,
	jobScheduler *scheduler.Scheduler,
) *Factory {
	return &Factory{
		settings:     settings,
		jobScheduler: jobScheduler,
	}
}

var apiRegex = regexp.MustCompile("phaas-.*-api")
var uiRegex = regexp.MustCompile("phaas-.*-ui")

func (this *Factory) BuildRepoController(ctx context.Context, path string, name string, dir os.DirEntry) Controller {
	if apiRegex.MatchString(name) {
		return &apiController{
			ctx:          ctx,
			appSettings:  this.settings,
			jobScheduler: this.jobScheduler,
			name:         name,
			path:         path,
			dir:          dir,
		}
	}
	return nil
}
