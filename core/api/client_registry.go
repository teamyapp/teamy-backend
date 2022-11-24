package api

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"google.golang.org/grpc"
)

type ClientRegistry struct {
	conn         *grpc.ClientConn
	taskClient   proto.TaskClient
	sprintClient proto.SprintClient
}

func (c *ClientRegistry) TaskClient() proto.TaskClient {
	if c.taskClient == nil {
		c.taskClient = proto.NewTaskClient(c.conn)
	}

	return c.taskClient
}

func (c *ClientRegistry) SprintClient() proto.SprintClient {
	if c.sprintClient == nil {
		c.sprintClient = proto.NewSprintClient(c.conn)
	}

	return c.sprintClient
}

func NewClientRegistry(dataCollector obs.DataCollector, connCfg rpc.ConnectionConfig, retry retry.Retry) (*ClientRegistry, error) {
	conn, err := rpc.NewClientConnection(dataCollector, connCfg, retry)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return &ClientRegistry{
		conn: conn,
	}, nil
}
