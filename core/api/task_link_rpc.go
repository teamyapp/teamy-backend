package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TaskLinkRPC struct {
	logger          telemetry.Logger
	taskLinkService service.TaskLink
	pbteamy.UnimplementedTaskLinkServer
}

var _ runner.Service = (*TaskLinkRPC)(nil)
var _ pbteamy.TaskLinkServer = (*TaskLinkRPC)(nil)

func (t TaskLinkRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterTaskLinkServer(server, t)
	})
	return nil
}

func (t TaskLinkRPC) CreateTaskLink(ct context.Context, in *pbteamy.CreateTaskLinkRequest) (*pbteamy.CreateTaskLinkResponse, error) {
	input := service.CreateTaskLinkInput{
		TaskID:       in.TaskId,
		Title:        in.Title,
		PreviewTitle: in.PreviewTitle,
		URL:          in.Url,
		IconURL:      in.IconUrl,
		IconHoverURL: in.IconHoverUrl,
	}

	taskLink, err := t.taskLinkService.CreateTaskLink(ct, input)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreateTaskLinkResponse{LinkId: taskLink.ID}, nil
}

func (t TaskLinkRPC) DeleteTaskLink(ct context.Context, in *pbteamy.DeleteTaskLinkRequest) (*emptypb.Empty, error) {
	_, err := t.taskLinkService.DeleteTaskLink(ct, in.LinkId)
	if err != nil {
		t.logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewTaskLinkRPC(logger telemetry.Logger, taskLinkService service.TaskLink) TaskLinkRPC {
	return TaskLinkRPC{
		logger:          logger,
		taskLinkService: taskLinkService,
	}
}
