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
)

type TaskLinkRPC struct {
	dataCollector   telemetry.DataCollector
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
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.CreateTaskLinkResponse{LinkId: taskLink.ID}, nil
}

func NewTaskLinkRPC(dataCollector telemetry.DataCollector, taskLinkService service.TaskLink) TaskLinkRPC {
	return TaskLinkRPC{
		dataCollector:   dataCollector,
		taskLinkService: taskLinkService,
	}
}
