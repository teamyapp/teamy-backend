package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TaskLinkRPC struct {
	logger          telemetry.Logger
	taskLinkService service.TaskLink
	proto.UnimplementedTaskLinkServer
}

var _ runner.Service = (*TaskLinkRPC)(nil)
var _ proto.TaskLinkServer = (*TaskLinkRPC)(nil)

func (t TaskLinkRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterTaskLinkServer(server, t)
	})
	return nil
}

func (t TaskLinkRPC) CreateTaskLink(ct context.Context, in *proto.CreateTaskLinkRequest) (*proto.CreateTaskLinkResponse, error) {
	input := service.CreateTaskLinkInput{
		TaskID:       in.TaskId,
		Title:        in.Title,
		URL:          in.Url,
		IconURL:      in.IconUrl,
		IconHoverURL: in.IconHoverUrl,
	}

	taskLink, err := t.taskLinkService.CreateTaskLink(ct, input)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.CreateTaskLinkResponse{LinkId: taskLink.ID}, nil
}

func (t TaskLinkRPC) DeleteTaskLink(ct context.Context, in *proto.DeleteTaskLinkRequest) (*emptypb.Empty, error) {
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
