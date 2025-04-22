package repo

import (
	"os"
	"phaas-localservices-ui/app"
	"regexp"
	"time"
)

type Controller interface {
	GetBasicDetails() BasicDetails
	GetStatus() (*Status, error)
	Start() error
	Stop() error
}

type Status struct {
	LastModified time.Time `json:"lastModified"`
	Branch       string    `json:"branch"`
	IsClean      bool      `json:"isClean"`
	Running      bool      `json:"running"`
}

type BasicDetails struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Factory struct {
	settings *app.Settings
}

func NewFactory(settings *app.Settings) *Factory {
	return &Factory{
		settings: settings,
	}
}

var apiRegex = regexp.MustCompile("phaas-.*-api")
var uiRegex = regexp.MustCompile("phaas-.*-ui")

func (this *Factory) BuildRepoController(path string, name string, dir os.DirEntry) Controller {
	if apiRegex.MatchString(name) {
		return &apiController{
			appSettings: this.settings,
			repoDetails: repoDetails{
				name: name,
				path: path,
				dir:  dir,
			},
		}
	}
	return nil
}
