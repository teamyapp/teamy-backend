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

type PhaseRPC struct {
	logger         telemetry.Logger
	phaseService   *service.Phase
	projectService *service.Project
	pbteamy.UnimplementedPhaseServiceServer
}

var _ runner.Service = (*PhaseRPC)(nil)
var _ pbteamy.PhaseServiceServer = (*PhaseRPC)(nil)

func (p PhaseRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterPhaseServiceServer(server, p)
	})

	return nil
}

func (p PhaseRPC) GetPhase(ct context.Context, req *pbteamy.GetPhaseRequest) (*pbteamy.GetPhaseResponse, error) {
	phase, err := p.phaseService.FindPhaseByID(ct, req.Id)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.GetPhaseResponse{
		Phase: &pbmessage.Phase{
			Id:              phase.ID,
			Name:            phase.Name,
			Status:          protoPhaseStatuses[phase.Status],
			ExpectedStartAt: toProtoTimePtr(&phase.ExpectedStartAt),
			ExpectedEndAt:   toProtoTimePtr(&phase.ExpectedEndAt),
			ActualStartAt:   toProtoTimePtr(phase.ActualStartAt),
			ActualEndAt:     toProtoTimePtr(phase.ActualEndAt),
			CreatedAt:       toProtoTimePtr(&phase.CreatedAt),
			UpdatedAt:       toProtoTimePtr(phase.UpdatedAt),
			CreatorUserId:   phase.CreatorID,
		},
	}, nil
}

func (p PhaseRPC) ListPhases(ct context.Context, req *pbteamy.ListPhasesRequest) (*pbteamy.ListPhasesResponse, error) {
	phases, err := p.projectService.FindPhasesByProjectID(ct, req.ProjectId)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var pbPhases []*pbmessage.Phase
	for _, phase := range phases {
		pbPhases = append(pbPhases, &pbmessage.Phase{
			Id:              phase.ID,
			Name:            phase.Name,
			Status:          protoPhaseStatuses[phase.Status],
			ExpectedStartAt: toProtoTimePtr(&phase.ExpectedStartAt),
			ExpectedEndAt:   toProtoTimePtr(&phase.ExpectedEndAt),
			ActualStartAt:   toProtoTimePtr(phase.ActualStartAt),
			ActualEndAt:     toProtoTimePtr(phase.ActualEndAt),
			CreatedAt:       toProtoTimePtr(&phase.CreatedAt),
			UpdatedAt:       toProtoTimePtr(phase.UpdatedAt),
			CreatorUserId:   phase.CreatorID,
		})
	}

	return &pbteamy.ListPhasesResponse{
		Phases: pbPhases,
	}, nil
}

func (p PhaseRPC) CreatePhase(ct context.Context, req *pbteamy.CreatePhaseRequest) (*pbteamy.CreatePhaseResponse, error) {
	expectedStartAt, err := fromProtoTime(req.ExpectedStartAt)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	expectedEndAt, err := fromProtoTime(req.ExpectedEndAt)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	createPhaseInput := service.CreatePhaseInput{
		Name:            req.Name,
		ExpectedStartAt: expectedStartAt,
		ExpectedEndAt:   expectedEndAt,
	}

	phase, err := p.phaseService.CreatePhase(ct, req.ProjectId, createPhaseInput)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreatePhaseResponse{
		Phase: &pbmessage.Phase{
			Id:              phase.ID,
			Name:            phase.Name,
			Status:          protoPhaseStatuses[phase.Status],
			ExpectedStartAt: toProtoTimePtr(&phase.ExpectedStartAt),
			ExpectedEndAt:   toProtoTimePtr(&phase.ExpectedEndAt),
			ActualStartAt:   toProtoTimePtr(phase.ActualStartAt),
			ActualEndAt:     toProtoTimePtr(phase.ActualEndAt),
			CreatedAt:       toProtoTimePtr(&phase.CreatedAt),
			UpdatedAt:       toProtoTimePtr(phase.UpdatedAt),
			CreatorUserId:   phase.CreatorID,
		},
	}, nil
}

func (p PhaseRPC) UpdatePhase(ct context.Context, req *pbteamy.UpdatePhaseRequest) (*pbteamy.UpdatePhaseResponse, error) {
	updatePhaseInput := service.UpdatePhaseInput{
		Name:            req.Name,
		ExpectedStartAt: fromProtoTimePtr(req.ExpectedStartAt),
		ExpectedEndAt:   fromProtoTimePtr(req.ExpectedEndAt),
		ActualStartAt:   fromProtoTimePtr(req.ActualStartAt),
		ActualEndAt:     fromProtoTimePtr(req.ActualEndAt),
		Status:          fromProtoPhaseStatusPtr(req.Status),
	}

	phase, err := p.phaseService.UpdatePhase(ct, req.Id, updatePhaseInput)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.UpdatePhaseResponse{
		Phase: &pbmessage.Phase{
			Id:              phase.ID,
			Name:            phase.Name,
			Status:          protoPhaseStatuses[phase.Status],
			ExpectedStartAt: toProtoTimePtr(&phase.ExpectedStartAt),
			ExpectedEndAt:   toProtoTimePtr(&phase.ExpectedEndAt),
			ActualStartAt:   toProtoTimePtr(phase.ActualStartAt),
			ActualEndAt:     toProtoTimePtr(phase.ActualEndAt),
			CreatedAt:       toProtoTimePtr(&phase.CreatedAt),
			UpdatedAt:       toProtoTimePtr(phase.UpdatedAt),
			CreatorUserId:   phase.CreatorID,
		},
	}, nil
}

func (p PhaseRPC) DeletePhase(ct context.Context, req *pbteamy.DeletePhaseRequest) (*emptypb.Empty, error) {
	_, err := p.phaseService.DeletePhase(ct, req.Id)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (p PhaseRPC) AddStoryToPhase(ct context.Context, req *pbteamy.AddStoryToPhaseRequest) (*emptypb.Empty, error) {
	_, err := p.phaseService.AddStoryToPhase(ct, req.PhaseId, req.StoryId)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (p PhaseRPC) RemoveStoryFromPhase(ct context.Context, req *pbteamy.RemoveStoryFromPhaseRequest) (*emptypb.Empty, error) {
	_, err := p.phaseService.RemoveStoryFromPhase(ct, req.PhaseId, req.StoryId)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (p PhaseRPC) AddStoriesToPhase(ct context.Context, req *pbteamy.AddStoriesToPhaseRequest) (*emptypb.Empty, error) {
	_, err := p.phaseService.AddStoriesToPhase(ct, req.PhaseId, req.StoryIds)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (p PhaseRPC) RemoveStoriesFromPhase(ct context.Context, req *pbteamy.RemoveStoriesFromPhaseRequest) (*emptypb.Empty, error) {
	_, err := p.phaseService.RemoveStoriesFromPhase(ct, req.PhaseId, req.StoryIds)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewPhaseRPC(logger telemetry.Logger, phaseService *service.Phase, projectService *service.Project) PhaseRPC {
	return PhaseRPC{
		logger:         logger,
		phaseService:   phaseService,
		projectService: projectService,
	}
}
