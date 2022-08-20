package api

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskRPC struct {
	taskService service.Task
	proto.UnimplementedTaskServer
}

var _ runner.Service = (*TaskRPC)(nil)
var _ proto.TaskServer = (*TaskRPC)(nil)

func (t TaskRPC) Start(runner *runner.ServiceRunner) error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterTaskServer(server, t)
	})
	return nil
}

func (t TaskRPC) FindTask(ct context.Context, req *proto.GetTaskRequest) (*proto.TaskMsg, error) {
	filter := &service.TaskFilter{
		TaskID: &req.TaskId,
	}
	tasks, err := t.taskService.FindTasks(ct, filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(tasks) < 1 {
		return &proto.TaskMsg{}, fmt.Errorf("task not found: taskID=%v", req.TaskId)
	}

	task := tasks[0]
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
	return &proto.CreateTaskResponse{TaskId: task.ID}, err
}

func (t TaskRPC) UpdateTask(ct context.Context, req *proto.UpdateTaskRequest) (*emptypb.Empty, error) {
	input := service.UpdateTaskInput{
		Goal:         req.Goal,
		Context:      req.Context,
		OwnerUserID:  req.OwnerUserId,
		OwningTeamID: req.OwningTeamId,
		Effort:       fromProtoDurationPtr(req.Effort),
		DueAt:        nil,
	}
	_, err := t.taskService.UpdateTask(ct, req.TaskId, input)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) DeleteTask(ct context.Context, req *proto.DeleteTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.DeleteTask(ct, req.TaskId)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) MoveTaskToUpcoming(ct context.Context, req *proto.MoveTaskToUpcomingRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToUpcoming(ct, req.TaskId, true)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) MoveTaskToInProgress(ct context.Context, req *proto.MoveTaskToInProgressRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToInProgress(ct, req.TaskId)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) MoveTaskToDelivered(ct context.Context, req *proto.MoveTaskToDeliveredRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToDelivered(ct, req.TaskId)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) MoveTaskToBlocked(ct context.Context, req *proto.MoveTaskToBlockedRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.MoveTaskToBlocked(ct, req.TaskId, req.Reason)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) AddAwaitForTask(ct context.Context, req *proto.AddAwaitForTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.AddAwaitForTask(ct, req.AwaitingTaskId, req.AwaitForTaskId)
	return &emptypb.Empty{}, err
}

func (t TaskRPC) RemoveAwaitForTask(ct context.Context, req *proto.RemoveAwaitForTaskRequest) (*emptypb.Empty, error) {
	_, err := t.taskService.RemoveAwaitForTask(ct, req.AwaitForTaskId, req.AwaitForTaskId)
	return &emptypb.Empty{}, err
}

func NewTaskRPC(taskService service.Task) TaskRPC {
	return TaskRPC{
		taskService: taskService,
	}
}
