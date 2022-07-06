package rpc

import (
	"fmt"

	"github.com/teamyapp/cloud/libs/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectionConfig struct {
	Host           string
	Port           int
	ShouldEncrypt  bool
	GetAccessToken func() string
}

func NewClientConnection(cfg ConnectionConfig) (*grpc.ClientConn, error) {
	var cred credentials.TransportCredentials
	if cfg.ShouldEncrypt {
		cred = credentials.NewTLS(nil)
	} else {
		cred = insecure.NewCredentials()
	}

	return grpc.Dial(
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(middleware.ClientWithGRPCIdentity(cfg.GetAccessToken)))
}
