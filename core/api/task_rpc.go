package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	"github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	pbmessage "github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskRPC struct {
	logger            telemetry.Logger
	taskService       service.Task
	taskLinkService   service.TaskLink
	attachmentService *service.Attachment
	pbteamy.UnimplementedTaskServiceServer
}

var _ runner.Service = (*TaskRPC)(nil)
var _ pbteamy.TaskServiceServer = (*TaskRPC)(nil)

func (t TaskRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterTaskServiceServer(server, t)
	})
	return nil
}

func (t TaskRPC) GetTask(ct context.Context, req *pbteamy.GetTaskRequest) (*pbteamy.GetTaskResponse, error) {
	task, err := t.taskService.FindTaskByID(ct, req.TaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.GetTaskResponse{
		Task: &pbmessage.Task{
			Id:               task.ID,
			Goal:             task.Goal,
			Context:          task.Context,
			Effort:           toProtoDurationPtr(task.Effort),
			Priority:         toProtoPriorityPtr(task.Priority),
			DueAt:            toProtoTimePtr(task.DueAt),
			Status:           protoTaskStatuses[task.Status],
			CreatedAt:        timestamppb.New(task.CreatedAt),
			UpdatedAt:        toProtoTimePtr(task.UpdatedAt),
			OwnerUserId:      task.OwnerUserID,
			OwningTeamId:     task.OwningTeamID,
			CreatorUserId:    task.CreatorUserID,
			CommentsThreadId: task.CommentsThreadID,
		},
	}, nil
}

func (t TaskRPC) GetAwaitForTasks(ct context.Context, req *pbteamy.GetAwaitForTasksRequest) (*pbteamy.GetAwaitForTasksResponse, error) {
	tasks, err := t.taskService.FindAwaitForTasks(ct, req.AwaitingTaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	taskMsgs := collect.Map(tasks, func(task entity.Task, _ int) *pbmessage.Task {
		return &pbmessage.Task{
			Id:               task.ID,
			Goal:             task.Goal,
			Context:          task.Context,
			Effort:           toProtoDurationPtr(task.Effort),
			Priority:         toProtoPriorityPtr(task.Priority),
			DueAt:            toProtoTimePtr(task.DueAt),
			Status:           protoTaskStatuses[task.Status],
			CreatedAt:        timestamppb.New(task.CreatedAt),
			UpdatedAt:        toProtoTimePtr(task.UpdatedAt),
			OwnerUserId:      task.OwnerUserID,
			OwningTeamId:     task.OwningTeamID,
			CreatorUserId:    task.CreatorUserID,
			CommentsThreadId: task.CommentsThreadID,
		}
	})

	return &pbteamy.GetAwaitForTasksResponse{Tasks: taskMsgs}, nil
}

func (t TaskRPC) CreateTask(ct context.Context, req *pbteamy.CreateTaskRequest) (*pbteamy.CreateTaskResponse, error) {
	input := service.CreateTaskInput{
		Goal:        req.Goal,
		Context:     req.Context,
		OwnerUserID: req.OwnerUserId,
		DueAt:       fromProtoTimePtr(req.DueAt),
	}
	task, err := t.taskService.CreateTask(ct, req.TeamId, input)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreateTaskResponse{TaskId: task.ID}, nil
}

func (t TaskRPC) UpdateTask(ct context.Context, req *pbteamy.UpdateTaskRequest) (*pbteamy.UpdateTaskResponse, error) {
	input := service.UpdateTaskInput{
		Goal:         req.Goal,
		Context:      req.Context,
		OwnerUserID:  req.OwnerUserId,
		OwningTeamID: req.OwningTeamId,
		Effort:       fromProtoDurationPtr(req.Effort),
		Priority:     fromProtoPriorityPtr(req.Priority),
		DueAt:        fromProtoTimePtr(req.DueAt),
	}
	updatedTask, err := t.taskService.UpdateTask(ct, req.TaskId, input)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.UpdateTaskResponse{
		Task: &message.Task{
			Id:               updatedTask.ID,
			Goal:             updatedTask.Goal,
			Context:          updatedTask.Context,
			Effort:           toProtoDurationPtr(updatedTask.Effort),
			Priority:         toProtoPriorityPtr(updatedTask.Priority),
			DueAt:            toProtoTimePtr(updatedTask.DueAt),
			Status:           protoTaskStatuses[updatedTask.Status],
			CreatedAt:        timestamppb.New(updatedTask.CreatedAt),
			UpdatedAt:        toProtoTimePtr(updatedTask.UpdatedAt),
			OwnerUserId:      updatedTask.OwnerUserID,
			OwningTeamId:     updatedTask.OwningTeamID,
			CreatorUserId:    updatedTask.CreatorUserID,
			CommentsThreadId: updatedTask.CommentsThreadID,
		},
	}, nil
}

func (t TaskRPC) DeleteTask(ct context.Context, req *pbteamy.DeleteTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.DeleteTask(ct, req.TaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToUpcoming(ct context.Context, req *pbteamy.MoveTaskToUpcomingRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToUpcoming(ct, req.TaskId, true)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToInProgress(ct context.Context, req *pbteamy.MoveTaskToInProgressRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToInProgress(ct, req.TaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToDelivered(ct context.Context, req *pbteamy.MoveTaskToDeliveredRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToDelivered(ct, req.TaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) MoveTaskToBlocked(ct context.Context, req *pbteamy.MoveTaskToBlockedRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToBlocked(ct, req.TaskId, req.Reason)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) AddAwaitForTask(ct context.Context, req *pbteamy.AddAwaitForTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.AddAwaitForTask(ct, req.AwaitingTaskId, req.AwaitForTaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (t TaskRPC) RemoveAwaitForTask(ct context.Context, req *pbteamy.RemoveAwaitForTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.RemoveAwaitForTask(ct, req.AwaitForTaskId, req.AwaitForTaskId)
	if err != nil {
		t.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewTaskRPC(logger telemetry.Logger, taskService service.Task, taskLinkService service.TaskLink, attachmentService *service.Attachment) TaskRPC {
	return TaskRPC{
		logger:            logger,
		taskService:       taskService,
		taskLinkService:   taskLinkService,
		attachmentService: attachmentService,
	}
}
