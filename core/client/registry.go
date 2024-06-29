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
	conn                  *grpc.ClientConn
	attachmentClient      pbteamy.AttachmentServiceClient
	messageClient         pbteamy.MessageServiceClient
	userClient            pbteamy.UserServiceClient
	teamClient            pbteamy.TeamServiceClient
	teamMemberGroupClient pbteamy.TeamMemberGroupServiceClient
	sprintClient          pbteamy.SprintServiceClient
	taskClient            pbteamy.TaskServiceClient
	taskLinkClient        pbteamy.TaskLinkServiceClient
}

func (r *Registry) AttachmentClient() pbteamy.AttachmentServiceClient {
	if r.attachmentClient == nil {
		r.attachmentClient = pbteamy.NewAttachmentServiceClient(r.conn)
	}

	return r.attachmentClient
}

func (r *Registry) MessageClient() pbteamy.MessageServiceClient {
	if r.messageClient == nil {
		r.messageClient = pbteamy.NewMessageServiceClient(r.conn)
	}

	return r.messageClient
}

func (r *Registry) TeamClient() pbteamy.TeamServiceClient {
	if r.teamClient == nil {
		r.teamClient = pbteamy.NewTeamServiceClient(r.conn)
	}

	return r.teamClient
}

func (r *Registry) TeamMemberGroupClient() pbteamy.TeamMemberGroupServiceClient {
	if r.teamMemberGroupClient == nil {
		r.teamMemberGroupClient = pbteamy.NewTeamMemberGroupServiceClient(r.conn)
	}

	return r.teamMemberGroupClient
}

func (r *Registry) UserClient() pbteamy.UserServiceClient {
	if r.userClient == nil {
		r.userClient = pbteamy.NewUserServiceClient(r.conn)
	}

	return r.userClient
}

func (r *Registry) SprintClient() pbteamy.SprintServiceClient {
	if r.sprintClient == nil {
		r.sprintClient = pbteamy.NewSprintServiceClient(r.conn)
	}

	return r.sprintClient
}

func (r *Registry) TaskClient() pbteamy.TaskServiceClient {
	if r.taskClient == nil {
		r.taskClient = pbteamy.NewTaskServiceClient(r.conn)
	}

	return r.taskClient
}

func (r *Registry) TaskLinkClient() pbteamy.TaskLinkServiceClient {
	if r.taskLinkClient == nil {
		r.taskLinkClient = pbteamy.NewTaskLinkServiceClient(r.conn)
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
