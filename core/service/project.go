package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type CreateProjectInput struct {
	Name            string
	ExpectedStartAt *time.Time
	ExpectedEndAt   *time.Time
	IconURL         *string
	Color           *string
}

type UpdateProjectInput struct {
	Name            string
	ExpectedStartAt *time.Time
	ActualStartAt   *time.Time
	ExpectedEndAt   *time.Time
	ActualEndAt     *time.Time
	IconURL         *string
	Color           *string
}

type Project struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	cloudClientRegistry     *client.Registry
	authorizer              client.Authorizer
	featureToggles          feature.Toggles
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	projectDao              dao.Project
	teamDao                 dao.Team
	phaseDao                dao.Phase
	storyDao                dao.Story
	projectPhaseRelationDao dao.ProjectPhaseRelation
	projectStoryRelationDao dao.ProjectStoryRelation
	userDao                 dao.User
	taskDao                 dao.Task
}

func (p *Project) FindProjects(ct context.Context, projectFilter *ProjectFilter) ([]entity.Project, *errs.Error) {
	var projects []entity.Project
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			projects, err = p.projectDao.FindProjectsWithTx(ct, tx)
			if err != nil {
				return err
			}

			if projectFilter != nil {
				projects = filterProjects(projects, *projectFilter)
			}

			return err
		})

	return projects, transactionErr
}

func (p *Project) FindProjectsByTeamID(ct context.Context, teamID uint64) ([]entity.Project, *errs.Error) {
	var projects []entity.Project
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			projects, err = p.projectDao.FindProjectsByTeamIDWithTx(ct, tx, teamID)
			return err
		})

	return projects, transactionErr
}

func (p *Project) FindStoriesByProjectID(ct context.Context, projectID uint64) ([]entity.Story, *errs.Error) {
	var stories []entity.Story
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			storyIDs, err := p.projectStoryRelationDao.FindStoryIDsByProjectIDWithTx(ct, tx, projectID)
			if err != nil {
				return err
			}

			stories, err = p.storyDao.FindStoriesByIDsWithTx(ct, tx, storyIDs)
			return err
		})

	return stories, transactionErr
}

func (p *Project) FindPhasesByProjectID(ct context.Context, projectID uint64) ([]entity.Phase, *errs.Error) {
	var phases []entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			phaseIDs, err := p.projectPhaseRelationDao.FindPhaseIDsByProjectIDWithTx(ct, tx, projectID)
			if err != nil {
				return err
			}

			phases, err = p.phaseDao.FindPhasesByIDsWithTx(ct, tx, phaseIDs)
			return err
		})

	return phases, transactionErr
}

func (p *Project) CreateProject(ct context.Context, teamID uint64, input CreateProjectInput) (entity.Project, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Project{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genProjectIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "projectID"}
	genProjectIDRes, rpcErr := p.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genProjectIDReq)
	if rpcErr != nil {
		return entity.Project{}, errs.FromGRPCErr(rpcErr)
	}

	project := entity.Project{
		ID:              genProjectIDRes.UniqueNumber,
		Name:            input.Name,
		ExpectedStartAt: input.ExpectedStartAt,
		ExpectedEndAt:   input.ExpectedEndAt,
		CreatorID:       userID,
		CreatedAt:       time.Now(),
		TeamID:          teamID,
		Color:           input.Color,
		IconURL:         input.IconURL,
	}

	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			_, err := p.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
			if err != nil {
				return err
			}

			return p.projectDao.CreateProject(ct, tx, project)
		})
	return project, transactionErr
}

func (p *Project) UpdateProject(ct context.Context, projectID uint64, input UpdateProjectInput) (entity.Project, *errs.Error) {
	var project entity.Project
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			project, err = p.projectDao.FindProjectByIDWithTx(ct, tx, projectID)
			if err != nil {
				return err
			}

			now := time.Now()
			project.Name = input.Name
			project.ExpectedStartAt = input.ExpectedStartAt
			project.ActualStartAt = input.ActualStartAt
			project.ExpectedEndAt = input.ExpectedEndAt
			project.ActualEndAt = input.ActualEndAt
			project.UpdatedAt = &now
			project.Color = input.Color
			project.IconURL = input.IconURL

			return p.projectDao.UpdateProject(ct, tx, project)
		})

	return project, transactionErr
}

func (p *Project) DeleteProject(ct context.Context, projectID uint64) (entity.Project, *errs.Error) {
	var project entity.Project
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			project, err = p.projectDao.FindProjectByIDWithTx(ct, tx, projectID)
			if err != nil {
				return err
			}

			phaseIDs, err := p.projectPhaseRelationDao.FindPhaseIDsByProjectIDWithTx(ct, tx, projectID)
			if err != nil {
				return err
			}

			err = p.projectPhaseRelationDao.DeleteProjectPhaseRelationsByProjectID(ct, tx, projectID)
			if err != nil {
				return err
			}

			err = p.projectStoryRelationDao.DeleteProjectStoryRelationsByProjectID(ct, tx, projectID)
			if err != nil {
				return err
			}

			err = p.phaseDao.DeletePhasesByIDs(ct, tx, phaseIDs)
			if err != nil {
				return err
			}

			return p.projectDao.DeleteProject(ct, tx, projectID)
		})

	return project, transactionErr
}

func NewProject(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	projectDao dao.Project,
	teamDao dao.Team,
	phaseDao dao.Phase,
	storyDao dao.Story,
	projectPhaseRelationDao dao.ProjectPhaseRelation,
	projectStoryRelationDao dao.ProjectStoryRelation,
	userDao dao.User,
	taskDao dao.Task,
) *Project {
	return &Project{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		cloudClientRegistry:     cloudClientRegistry,
		authorizer:              authorizer,
		featureToggles:          featureToggles,
		transactionFactory:      transactionFactory,
		stateSyncer:             stateSyncer,
		projectDao:              projectDao,
		teamDao:                 teamDao,
		phaseDao:                phaseDao,
		storyDao:                storyDao,
		projectPhaseRelationDao: projectPhaseRelationDao,
		projectStoryRelationDao: projectStoryRelationDao,
		userDao:                 userDao,
		taskDao:                 taskDao,
	}
}
