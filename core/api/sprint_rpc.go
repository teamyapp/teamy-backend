package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	pbmessage "github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SprintRPC struct {
	logger        telemetry.Logger
	sprintService service.Sprint
	pbteamy.UnimplementedSprintServiceServer
}

var _ runner.Service = (*SprintRPC)(nil)
var _ pbteamy.SprintServiceServer = (*SprintRPC)(nil)

func (s SprintRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterSprintServiceServer(server, s)
	})
	return nil
}

func (s SprintRPC) GetActiveSprint(ct context.Context, req *pbteamy.GetActiveSprintRequest) (*pbteamy.GetActiveSprintResponse, error) {
	sprint, err := s.sprintService.GetActiveSprint(ct, req.TeamId)
	if err != nil {
		s.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.GetActiveSprintResponse{
		Sprint: &pbmessage.Sprint{
			Id:           sprint.ID,
			StartAt:      timestamppb.New(sprint.StartAt),
			EndAt:        timestamppb.New(sprint.EndAt),
			CreatedAt:    timestamppb.New(sprint.CreatedAt),
			OwningTeamId: sprint.OwningTeamID,
		},
	}, nil
}

func (s SprintRPC) AddTaskToSprint(ct context.Context, req *pbteamy.AddTaskToSprintRequest) (*emptypb.Empty, error) {
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
