package api

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	pbmessage "github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StoryRPC struct {
	logger       telemetry.Logger
	storyService *service.Story
	pbteamy.UnimplementedStoryServiceServer
}

var _ runner.Service = (*StoryRPC)(nil)
var _ pbteamy.StoryServiceServer = (*StoryRPC)(nil)

func (s StoryRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterStoryServiceServer(server, s)
	})

	return nil
}

func (s StoryRPC) GetStory(ct context.Context, req *pbteamy.GetStoryRequest) (*pbteamy.GetStoryResponse, error) {
	story, err := s.storyService.FindStoryByID(ct, req.Id)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.GetStoryResponse{
		Story: &pbmessage.Story{
			Id:            story.ID,
			Name:          story.Name,
			OwnerUserId:   story.OwnerID,
			IsPlanned:     story.IsPlanned,
			Status:        protoStoryStatuses[story.Status],
			Priority:      toProtoPriorityPtr(story.Priority),
			CreatorUserId: story.CreatorID,
			CreatedAt:     toProtoTimePtr(&story.CreatedAt),
			UpdatedAt:     toProtoTimePtr(story.UpdatedAt),
		},
	}, nil
}

func (s StoryRPC) GetTasksByStory(ct context.Context, req *pbteamy.GetTasksByStoryRequest) (*pbteamy.GetTasksByStoryResponse, error) {
	tasks, err := s.storyService.GetTasksByStory(ct, req.StoryId)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var pbTasks []*pbmessage.Task
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pbmessage.Task{
			Id:               task.ID,
			Goal:             task.Goal,
			Context:          task.Context,
			Effort:           toProtoDurationPtr(task.Effort),
			Priority:         toProtoPriorityPtr(task.Priority),
			DueAt:            toProtoTimePtr(task.DueAt),
			Status:           protoTaskStatuses[task.Status],
			CreatedAt:        toProtoTimePtr(&task.CreatedAt),
			UpdatedAt:        toProtoTimePtr(task.UpdatedAt),
			OwnerUserId:      task.OwnerUserID,
			OwningTeamId:     task.OwningTeamID,
			CreatorUserId:    task.CreatorUserID,
			CommentsThreadId: task.CommentsThreadID,
		})
	}

	return &pbteamy.GetTasksByStoryResponse{
		Tasks: pbTasks,
	}, nil
}

func (s StoryRPC) ListStories(ct context.Context, req *pbteamy.ListStoriesRequest) (*pbteamy.ListStoriesResponse, error) {
	stories, err := s.storyService.ListStories(ct, service.StoryQuery{
		ProjectID: req.ProjectId,
		PhaseID:   req.PhaseId,
	})

	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var pbStories []*pbmessage.Story
	for _, story := range stories {
		pbStories = append(pbStories, &pbmessage.Story{
			Id:            story.ID,
			Name:          story.Name,
			OwnerUserId:   story.OwnerID,
			IsPlanned:     story.IsPlanned,
			Status:        protoStoryStatuses[story.Status],
			Priority:      toProtoPriorityPtr(story.Priority),
			CreatorUserId: story.CreatorID,
			CreatedAt:     toProtoTimePtr(&story.CreatedAt),
			UpdatedAt:     toProtoTimePtr(story.UpdatedAt),
		})
	}

	return &pbteamy.ListStoriesResponse{
		Stories: pbStories,
	}, nil
}

func (s StoryRPC) CreateStory(ct context.Context, req *pbteamy.CreateStoryRequest) (*pbteamy.CreateStoryResponse, error) {
	createStoryInput := service.CreateStoryInput{
		Name:     req.Name,
		Status:   storyStatuses[req.Status],
		OwnerID:  req.OwnerUserId,
		Priority: fromProtoPriorityPtr(req.Priority),
	}

	story, err := s.storyService.CreateStory(ct, req.ProjectId, createStoryInput)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreateStoryResponse{
		Story: &pbmessage.Story{
			Id:            story.ID,
			Name:          story.Name,
			OwnerUserId:   story.OwnerID,
			IsPlanned:     story.IsPlanned,
			Status:        protoStoryStatuses[story.Status],
			Priority:      toProtoPriorityPtr(story.Priority),
			CreatorUserId: story.CreatorID,
			CreatedAt:     toProtoTimePtr(&story.CreatedAt),
			UpdatedAt:     toProtoTimePtr(story.UpdatedAt),
		},
	}, nil
}

func (s StoryRPC) UpdateStory(ct context.Context, req *pbteamy.UpdateStoryRequest) (*pbteamy.UpdateStoryResponse, error) {
	updateStoryInput := service.UpdateStoryInput{
		Name:     req.Name,
		Status:   fromProtoStoryStatusPtr(req.Status),
		OwnerID:  req.OwnerUserId,
		Priority: fromProtoPriorityPtr(req.Priority),
	}

	story, err := s.storyService.UpdateStory(ct, req.Id, updateStoryInput)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.UpdateStoryResponse{
		Story: &pbmessage.Story{
			Id:            story.ID,
			Name:          story.Name,
			OwnerUserId:   story.OwnerID,
			IsPlanned:     story.IsPlanned,
			Status:        protoStoryStatuses[story.Status],
			Priority:      toProtoPriorityPtr(story.Priority),
			CreatorUserId: story.CreatorID,
			CreatedAt:     toProtoTimePtr(&story.CreatedAt),
			UpdatedAt:     toProtoTimePtr(story.UpdatedAt),
		},
	}, nil
}

func (s StoryRPC) DeleteStory(ct context.Context, req *pbteamy.DeleteStoryRequest) (*emptypb.Empty, error) {
	_, err := s.storyService.DeleteStory(ct, req.Id)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (s StoryRPC) AddTaskToStory(ct context.Context, req *pbteamy.AddTaskToStoryRequest) (*emptypb.Empty, error) {
	_, err := s.storyService.AddTaskToStory(ct, req.StoryId, req.TaskId)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (s StoryRPC) RemoveTaskFromStory(ct context.Context, req *pbteamy.RemoveTaskFromStoryRequest) (*emptypb.Empty, error) {
	_, err := s.storyService.RemoveTaskFromStory(ct, req.StoryId, req.TaskId)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (s StoryRPC) AddTasksToStory(ct context.Context, req *pbteamy.AddTasksToStoryRequest) (*emptypb.Empty, error) {
	_, err := s.storyService.AddTasksToStory(ct, req.StoryId, req.TaskIds)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (s StoryRPC) RemoveTasksFromStory(ct context.Context, req *pbteamy.RemoveTasksFromStoryRequest) (*emptypb.Empty, error) {
	_, err := s.storyService.RemoveTasksFromStory(ct, req.StoryId, req.TaskIds)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewStoryRPC(
	logger telemetry.Logger,
	storyService *service.Story,
) StoryRPC {
	return StoryRPC{
		logger:       logger,
		storyService: storyService,
	}
}
