package rpc

import (
	"fmt"

	"github.com/teamyapp/cloud/app/api/rpc/proto"
	"github.com/teamyapp/cloud/app/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CloudAPIClient struct {
	gRPCConn        *grpc.ClientConn
	generatorClient *proto.GeneratorClient
}

func (c *CloudAPIClient) GeneratorClient() proto.GeneratorClient {
	if c.generatorClient == nil {
		client := proto.NewGeneratorClient(c.gRPCConn)
		c.generatorClient = &client
	}

	return *c.generatorClient
}

func (c *CloudAPIClient) Close() error {
	return c.gRPCConn.Close()
}

func NewCloudAPIClient(cfg config.CloudAPIClient) (*CloudAPIClient, error) {
	var opts grpc.DialOption
	if cfg.ShouldEncrypt {
		opts = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	} else {
		opts = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), opts)
	if err != nil {
		return nil, err
	}

	return &CloudAPIClient{
		gRPCConn: conn,
	}, nil
}
