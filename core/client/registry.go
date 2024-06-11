package client

import (
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	"google.golang.org/grpc"
)

type Registry struct {
	conn           *grpc.ClientConn
	teamClient     pbteamy.TeamClient
	sprintClient   pbteamy.SprintClient
	taskClient     pbteamy.TaskClient
	taskLinkClient pbteamy.TaskLinkClient
}

func (r *Registry) TeamClient() pbteamy.TeamClient {
	if r.teamClient == nil {
		r.teamClient = pbteamy.NewTeamClient(r.conn)
	}

	return r.teamClient
}

func (r *Registry) SprintClient() pbteamy.SprintClient {
	if r.sprintClient == nil {
		r.sprintClient = pbteamy.NewSprintClient(r.conn)
	}

	return r.sprintClient
}

func (r *Registry) TaskClient() pbteamy.TaskClient {
	if r.taskClient == nil {
		r.taskClient = pbteamy.NewTaskClient(r.conn)
	}

	return r.taskClient
}

func (r *Registry) TaskLinkClient() pbteamy.TaskLinkClient {
	if r.taskLinkClient == nil {
		r.taskLinkClient = pbteamy.NewTaskLinkClient(r.conn)
	}

	return r.taskLinkClient
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
