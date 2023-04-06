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

type SprintRPC struct {
	logger        telemetry.Logger
	sprintService service.Sprint
	proto.UnimplementedSprintServer
}

var _ runner.Service = (*SprintRPC)(nil)
var _ proto.SprintServer = (*SprintRPC)(nil)

func (s SprintRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterSprintServer(server, s)
	})
	return nil
}

func (s SprintRPC) GetCurrentSprint(ct context.Context, req *proto.GetCurrentSprintRequest) (*proto.SprintMsg, error) {
	sprint, err := s.sprintService.FindCurrentSprint(ct, req.TeamId)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
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
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewSprintRPC(logger telemetry.Logger, sprintService service.Sprint) SprintRPC {
	return SprintRPC{
		logger:        logger,
		sprintService: sprintService,
	}
}
