package runner

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/teamy-backend/infras/web"
)

type ServiceRunnerConfig struct {
	WebServerPort int `envconfig:"SERVICE_RUNNER_WEB_SERVER_PORT" default:"9001"`
}

func ServiceRunnerConfigFromEnv() (ServiceRunnerConfig, error) {
	cfg := ServiceRunnerConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return ServiceRunnerConfig{}, err
	}

	return cfg, nil
}

type ServiceRunner struct {
	config   ServiceRunnerConfig
	router   *mux.Router
	services []Service
}

func (s *ServiceRunner) Start() error {
	for _, service := range s.services {
		err := service.Start(s)
		if err != nil {
			return err
		}
	}

	err := web.StartWebServer(s.router, s.config.WebServerPort)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceRunner) RegisterWebRoutes(routes []web.Route) {
	for _, route := range routes {
		s.router.HandleFunc(route.Path, route.HandlerFunc).Methods(route.Method)
	}
}

func NewServiceRunner(config ServiceRunnerConfig, services []Service) ServiceRunner {
	return ServiceRunner{
		config:   config,
		router:   mux.NewRouter(),
		services: services,
	}
}
