package runner

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

type ServiceRunnerConfig struct {
	WebServerPort  int `envconfig:"SERVICE_RUNNER_WEB_SERVER_PORT" default:"9011"`
	GRPCServerPort int `envconfig:"SERVICE_RUNNER_GRPC_SERVER_PORT" default:"9012"`
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
	config     ServiceRunnerConfig
	webRouter  *mux.Router
	gRPCServer *grpc.Server
	services   []Service
}

func (s *ServiceRunner) Start() {
	for _, service := range s.services {
		err := service.Start(s)
		if err != nil {
			panic(err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Service runner web server started at port %d\n", s.config.WebServerPort)
		err := web.StartWebServer(s.webRouter, s.config.WebServerPort)
		if err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", fmt.Sprintf(fmt.Sprintf(":%d", s.config.GRPCServerPort)))
		if err != nil {
			panic(err)
		}

		log.Printf("Service runner gRPC server started at port %d\n", s.config.GRPCServerPort)
		err = s.gRPCServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}

func (s *ServiceRunner) RegisterWebRoutes(routes []web.Route) {
	for _, route := range routes {
		s.webRouter.HandleFunc(route.Path, route.HandlerFunc).Methods(route.Method)
	}
}

func (s *ServiceRunner) WithGRPCServer(withGRPCServer func(server *grpc.Server)) {
	withGRPCServer(s.gRPCServer)
}

func NewServiceRunner(config ServiceRunnerConfig, services []Service) ServiceRunner {
	return ServiceRunner{
		config:     config,
		webRouter:  mux.NewRouter(),
		gRPCServer: grpc.NewServer(),
		services:   services,
	}
}
