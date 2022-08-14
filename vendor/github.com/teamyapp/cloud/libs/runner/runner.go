package runner

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/middleware"
	"google.golang.org/grpc"
)

type WebRoute struct {
	Path        string
	Method      string
	HandlerFunc http.HandlerFunc
}

type ServiceRunnerConfig struct {
	WebServerPort       int    `envconfig:"SERVICE_RUNNER_WEB_SERVER_PORT" default:"9011"`
	GRPCServerPort      int    `envconfig:"SERVICE_RUNNER_GRPC_SERVER_PORT" default:"9012"`
	IdentityAPIEndpoint string `envconfig:"SERVICE_RUNNER_IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
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
		s.startWebServer()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.startGRPCServer()
	}()
	wg.Wait()
}

func (s *ServiceRunner) startWebServer() {
	log.Printf("Service runner Web server started at port %d\n", s.config.WebServerPort)
	serveMux := http.NewServeMux()
	handlerFunc := middleware.EnableCORS(
		middleware.LogWebRequest(
			middleware.ServerWithWebIdentity(
				s.config.IdentityAPIEndpoint,
				middleware.ServerWithWebSocketIdentity(s.config.IdentityAPIEndpoint,
					s.webRouter.ServeHTTP))))
	serveMux.HandleFunc("/", handlerFunc)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.WebServerPort), serveMux); err != nil {
		panic(err)
	}
}

func (s *ServiceRunner) startGRPCServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(fmt.Sprintf(":%d", s.config.GRPCServerPort)))
	if err != nil {
		panic(err)
	}

	log.Printf("Service runner gRPC server started at port %d\n", s.config.GRPCServerPort)
	err = s.gRPCServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}

func (s *ServiceRunner) RegisterWebRoutes(routes []WebRoute) {
	for _, route := range routes {
		s.webRouter.HandleFunc(route.Path, route.HandlerFunc).Methods(route.Method)
	}
}

func (s *ServiceRunner) WithGRPCServer(withGRPCServer func(server *grpc.Server)) {
	withGRPCServer(s.gRPCServer)
}

func NewServiceRunner(config ServiceRunnerConfig, services []Service) ServiceRunner {
	return ServiceRunner{
		config:    config,
		webRouter: mux.NewRouter(),
		gRPCServer: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.LogGRPCRequest,
				middleware.ServerWithGRPCIdentity(config.IdentityAPIEndpoint),
			)),
		services: services,
	}
}
