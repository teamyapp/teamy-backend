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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskRPC struct {
	dataCollector telemetry.DataCollector
	taskService   service.Task
	proto.UnimplementedTaskServer
}

var _ runner.Service = (*TaskRPC)(nil)
var _ proto.TaskServer = (*TaskRPC)(nil)

func (t TaskRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterTaskServer(server, t)
	})
	return nil
}

func (t TaskRPC) GetTask(ct context.Context, req *proto.GetTaskRequest) (*proto.TaskMsg, error) {
	task, err := t.taskService.FindTaskByID(ct, req.TaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.TaskMsg{
		TaskId:          task.ID,
		Goal:            task.Goal,
		Context:         task.Context,
		Effort:          toProtoDurationPtr(task.Effort),
		DueAt:           toProtoTimePtr(task.DueAt),
		Status:          protoTaskStatuses[task.Status],
		CreatedAt:       timestamppb.New(task.CreatedAt),
		UpdatedAt:       toProtoTimePtr(task.UpdatedAt),
		OwnerUserId:     task.OwnerUserID,
		OwningTeamId:    task.OwningTeamID,
		CreatorUserId:   task.CreatorUserID,
		CommentThreadId: task.CommentsThreadID,
	}, nil
}

func (t TaskRPC) CreateTask(ct context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	input := service.CreateTaskInput{
		Goal:        req.Goal,
		Context:     req.Context,
		OwnerUserID: req.OwnerUserId,
		DueAt:       fromProtoTimePtr(req.DueAt),
	}
	task, err := t.taskService.CreateTask(ct, req.TeamId, input)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.CreateTaskResponse{TaskId: task.ID}, nil
}

func (t TaskRPC) UpdateTask(ct context.Context, req *proto.UpdateTaskRequest) (*emptypb.Empty, error) {
	input := service.UpdateTaskInput{
		Goal:         req.Goal,
		Context:      req.Context,
		OwnerUserID:  req.OwnerUserId,
		OwningTeamID: req.OwningTeamId,
		Effort:       fromProtoDurationPtr(req.Effort),
		DueAt:        fromProtoTimePtr(req.DueAt),
	}
	_, err := t.taskService.UpdateTask(ct, req.TaskId, input)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) DeleteTask(ct context.Context, req *proto.DeleteTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.DeleteTask(ct, req.TaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToUpcoming(ct context.Context, req *proto.MoveTaskToUpcomingRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToUpcoming(ct, req.TaskId, true)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToInProgress(ct context.Context, req *proto.MoveTaskToInProgressRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToInProgress(ct, req.TaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToDelivered(ct context.Context, req *proto.MoveTaskToDeliveredRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToDelivered(ct, req.TaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToBlocked(ct context.Context, req *proto.MoveTaskToBlockedRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToBlocked(ct, req.TaskId, req.Reason)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) AddAwaitForTask(ct context.Context, req *proto.AddAwaitForTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.AddAwaitForTask(ct, req.AwaitingTaskId, req.AwaitForTaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) RemoveAwaitForTask(ct context.Context, req *proto.RemoveAwaitForTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.RemoveAwaitForTask(ct, req.AwaitForTaskId, req.AwaitForTaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewTaskRPC(dataCollector telemetry.DataCollector, taskService service.Task) TaskRPC {
	return TaskRPC{
		dataCollector: dataCollector,
		taskService:   taskService,
	}
}
