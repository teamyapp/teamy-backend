package client

import (
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"google.golang.org/grpc"
)

type Registry struct {
	conn           *grpc.ClientConn
	taskClient     proto.TaskClient
	taskLinkClient proto.TaskLinkClient
	sprintClient   proto.SprintClient
}

func (r *Registry) TaskClient() proto.TaskClient {
	if r.taskClient == nil {
		r.taskClient = proto.NewTaskClient(r.conn)
	}

	return r.taskClient
}

func (r *Registry) TaskLinkClient() proto.TaskLinkClient {
	if r.taskLinkClient == nil {
		r.taskLinkClient = proto.NewTaskLinkClient(r.conn)
	}

	return r.taskLinkClient
}

func (r *Registry) SprintClient() proto.SprintClient {
	if r.sprintClient == nil {
		r.sprintClient = proto.NewSprintClient(r.conn)
	}

	return r.sprintClient
}

func NewRegistry(
	logger telemetry.Logger,
	network network.Network,
	clientGRPCMetrics middleware.ClientGRPCMetrics,
	connCfg rpc.ConnectionConfig,
	makeRetry func() retry.Retry,
) (*Registry, *errs.Error) {
	conn, err := rpc.NewClientConnection(logger, network, clientGRPCMetrics, connCfg, makeRetry)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return &Registry{
		conn: conn,
	}, nil
}
