package rpc

import (
	"google.golang.org/grpc"
)

type Service interface {
	registerServer(server *grpc.Server)
}

func NewServer(services []Service) *grpc.Server {
	s := grpc.NewServer()
	for _, service := range services {
		service.registerServer(s)
	}

	return s
}
