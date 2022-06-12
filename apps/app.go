package apps

import (
	"github.com/gorilla/mux"
	"github.com/teamyapp/teamy-backend/apps/config"
)

type App interface {
	init(runner *AppRunner) error
}

type AppRunner struct {
	config config.AppRunnerConfig
	router *mux.Router
	apps   []App
}

func (a *AppRunner) Start() error {
	for _, app := range a.apps {
		err := app.init(a)
		if err != nil {
			return err
		}
	}

	err := startWebServer(a.router, a.config.WebAPIPort)
	if err != nil {
		return err
	}

	return nil
}

func (a AppRunner) registerRoute(route Route) {
	a.router.HandleFunc(route.Path, route.HandlerFunc).Methods(route.Method)
}

func NewAppRunner(config config.AppRunnerConfig, apps []App) AppRunner {
	return AppRunner{
		config: config,
		router: mux.NewRouter(),
		apps:   apps,
	}
}
