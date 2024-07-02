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

type ProjectRPC struct {
	logger         telemetry.Logger
	projectService *service.Project
	pbteamy.UnimplementedProjectServiceServer
}

var _ runner.Service = (*ProjectRPC)(nil)
var _ pbteamy.ProjectServiceServer = (*ProjectRPC)(nil)

func (p ProjectRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterProjectServiceServer(server, p)
	})
	return nil
}

func (p ProjectRPC) GetProject(ct context.Context, req *pbteamy.GetProjectRequest) (*pbteamy.GetProjectResponse, error) {
	project, err := p.projectService.FindProjectByID(ct, req.Id)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.GetProjectResponse{
		Project: &pbmessage.Project{
			Id:                  project.ID,
			Name:                project.Name,
			ExpectedStartAt:     toProtoTimePtr(project.ExpectedStartAt),
			ExpectedEndAt:       toProtoTimePtr(project.ExpectedEndAt),
			ActualStartAt:       toProtoTimePtr(project.ActualStartAt),
			ActualEndAt:         toProtoTimePtr(project.ActualEndAt),
			IconUrl:             project.IconURL,
			Color:               project.Color,
			CreatedAt:           toProtoTimePtr(&project.CreatedAt),
			UpdatedAt:           toProtoTimePtr(project.UpdatedAt),
			CreatorUserId:       project.CreatorID,
			TeamId:              project.TeamID,
			TotalPhaseCount:     int32(project.TotalPhaseCount),
			CompletedPhaseCount: int32(project.CompletedPhaseCount),
		},
	}, nil
}

func (p ProjectRPC) ListProjects(ct context.Context, req *pbteamy.ListProjectsRequest) (*pbteamy.ListProjectsResponse, error) {
	projects, err := p.projectService.FindProjectsByTeamID(ct,
		req.TeamId,
	)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var pbProjects []*pbmessage.Project
	for _, project := range projects {
		pbProjects = append(pbProjects, &pbmessage.Project{
			Id:                  project.ID,
			Name:                project.Name,
			ExpectedStartAt:     toProtoTimePtr(project.ExpectedStartAt),
			ExpectedEndAt:       toProtoTimePtr(project.ExpectedEndAt),
			ActualStartAt:       toProtoTimePtr(project.ActualStartAt),
			ActualEndAt:         toProtoTimePtr(project.ActualEndAt),
			IconUrl:             project.IconURL,
			Color:               project.Color,
			CreatedAt:           toProtoTimePtr(&project.CreatedAt),
			UpdatedAt:           toProtoTimePtr(project.UpdatedAt),
			CreatorUserId:       project.CreatorID,
			TeamId:              project.TeamID,
			TotalPhaseCount:     int32(project.TotalPhaseCount),
			CompletedPhaseCount: int32(project.CompletedPhaseCount),
		})
	}

	return &pbteamy.ListProjectsResponse{
		Projects: pbProjects,
	}, nil
}

func (p ProjectRPC) CreateProject(ct context.Context, req *pbteamy.CreateProjectRequest) (*pbteamy.CreateProjectResponse, error) {
	project, err := p.projectService.CreateProject(ct, req.TeamId, service.CreateProjectInput{
		Name:            req.Name,
		ExpectedStartAt: fromProtoTimePtr(req.ExpectedStartAt),
		ExpectedEndAt:   fromProtoTimePtr(req.ExpectedEndAt),
		Color:           req.Color,
		IconURL:         req.IconUrl,
	})

	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreateProjectResponse{
		Project: &pbmessage.Project{
			Id:                  project.ID,
			Name:                project.Name,
			ExpectedStartAt:     toProtoTimePtr(project.ExpectedStartAt),
			ExpectedEndAt:       toProtoTimePtr(project.ExpectedEndAt),
			ActualStartAt:       toProtoTimePtr(project.ActualStartAt),
			ActualEndAt:         toProtoTimePtr(project.ActualEndAt),
			IconUrl:             project.IconURL,
			Color:               project.Color,
			CreatedAt:           toProtoTimePtr(&project.CreatedAt),
			UpdatedAt:           toProtoTimePtr(project.UpdatedAt),
			CreatorUserId:       project.CreatorID,
			TeamId:              project.TeamID,
			TotalPhaseCount:     int32(project.TotalPhaseCount),
			CompletedPhaseCount: int32(project.CompletedPhaseCount),
		},
	}, nil
}

func (p ProjectRPC) UpdateProject(ct context.Context, req *pbteamy.UpdateProjectRequest) (*pbteamy.UpdateProjectResponse, error) {
	project, err := p.projectService.UpdateProject(ct, req.Id, service.UpdateProjectInput{
		Name:            req.Name,
		ExpectedStartAt: fromProtoTimePtr(req.ExpectedStartAt),
		ExpectedEndAt:   fromProtoTimePtr(req.ExpectedEndAt),
		ActualStartAt:   fromProtoTimePtr(req.ActualStartAt),
		ActualEndAt:     fromProtoTimePtr(req.ActualEndAt),
		Color:           req.Color,
		IconURL:         req.IconUrl,
	})

	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.UpdateProjectResponse{
		Project: &pbmessage.Project{
			Id:                  project.ID,
			Name:                project.Name,
			ExpectedStartAt:     toProtoTimePtr(project.ExpectedStartAt),
			ExpectedEndAt:       toProtoTimePtr(project.ExpectedEndAt),
			ActualStartAt:       toProtoTimePtr(project.ActualStartAt),
			ActualEndAt:         toProtoTimePtr(project.ActualEndAt),
			IconUrl:             project.IconURL,
			Color:               project.Color,
			CreatedAt:           toProtoTimePtr(&project.CreatedAt),
			UpdatedAt:           toProtoTimePtr(project.UpdatedAt),
			CreatorUserId:       project.CreatorID,
			TeamId:              project.TeamID,
			TotalPhaseCount:     int32(project.TotalPhaseCount),
			CompletedPhaseCount: int32(project.CompletedPhaseCount),
		},
	}, nil
}

func (p ProjectRPC) DeleteProject(ct context.Context, req *pbteamy.DeleteProjectRequest) (*emptypb.Empty, error) {
	_, err := p.projectService.DeleteProject(ct, req.Id)
	if err != nil {
		p.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewProjectRPC(logger telemetry.Logger, projectService *service.Project) ProjectRPC {
	return ProjectRPC{
		logger:         logger,
		projectService: projectService,
	}
}
