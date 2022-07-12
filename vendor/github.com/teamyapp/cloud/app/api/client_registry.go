package api

import (
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/rpc"
	"google.golang.org/grpc"
)

type ClientRegistry struct {
	conn            *grpc.ClientConn
	generatorClient proto.GeneratorClient
	identityClient  proto.IdentityClient
}

func (c *ClientRegistry) GeneratorClient() proto.GeneratorClient {
	if c.generatorClient == nil {
		c.generatorClient = proto.NewGeneratorClient(c.conn)
	}

	return c.generatorClient
}

func (c *ClientRegistry) IdentityClient() proto.IdentityClient {
	if c.identityClient == nil {
		c.identityClient = proto.NewIdentityClient(c.conn)
	}

	return c.identityClient
}

func NewClientRegistry(connCfg rpc.ConnectionConfig) (*ClientRegistry, error) {
	conn, err := rpc.NewClientConnection(connCfg)
	if err != nil {
		return nil, err
	}

	return &ClientRegistry{
		conn: conn,
	}, nil
}
