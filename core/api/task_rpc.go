package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
)

type TaskRPC struct {
	taskService service.Task
	proto.UnimplementedTaskServer
}

var _ runner.Service = (*TaskRPC)(nil)
var _ proto.UnsafeTaskServer = (*TaskRPC)(nil)

func (t TaskRPC) Start(runner *runner.ServiceRunner) error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterTaskServer(server, t)
	})
	return nil
}

func (t TaskRPC) CreateTask(ct context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	input := service.CreateTaskInput{
		Goal:        req.Goal,
		Context:     req.Context,
		OwnerUserID: req.OwnerUserId,
		DueAt:       fromProtoTimePtr(req.DueAt),
	}
	task, err := t.taskService.CreateTask(ct, req.TeamId, input)
	return &proto.CreateTaskResponse{TaskId: task.ID}, err
}

func NewTaskRPC(taskService service.Task) TaskRPC {
	return TaskRPC{
		taskService: taskService,
	}
}
