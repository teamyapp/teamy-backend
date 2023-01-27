package api

import (
	"context"
	_ "embed"
	"errors"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SprintRPC struct {
	dataCollector obs.DataCollector
	sprintService service.Sprint
	proto.UnimplementedSprintServer
}

var _ runner.Service = (*SprintRPC)(nil)
var _ proto.SprintServer = (*SprintRPC)(nil)

func (s SprintRPC) Start(runner *runner.ServiceRunner) error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterSprintServer(server, s)
	})
	return nil
}

func (s SprintRPC) GetCurrentSprint(ct context.Context, req *proto.GetCurrentSprintRequest) (*proto.SprintMsg, error) {
	sprint, err := s.sprintService.FindCurrentSprint(ct, req.TeamId)
	if err != nil {
		if errors.As(err, &service.ErrorNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return &proto.SprintMsg{
		Id:           sprint.ID,
		StartAt:      timestamppb.New(sprint.StartAt),
		EndAt:        timestamppb.New(sprint.EndAt),
		CreatedAt:    timestamppb.New(sprint.CreatedAt),
		OwningTeamId: sprint.OwningTeamID,
	}, nil
}

func (s SprintRPC) AddTaskToSprint(ct context.Context, req *proto.AddTaskToSprintRequest) (*emptypb.Empty, error) {
	_, err := s.sprintService.AddTaskToSprint(ct, req.SprintId, req.TaskId)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return &emptypb.Empty{}, err
}

func NewSprintRPC(dataCollector obs.DataCollector, sprintService service.Sprint) SprintRPC {
	return SprintRPC{
		dataCollector: dataCollector,
		sprintService: sprintService,
	}
}
