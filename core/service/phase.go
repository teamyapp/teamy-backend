package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type CreatePhaseInput struct {
	Name            string
	ExpectedStartAt time.Time
	ExpectedEndAt   time.Time
}

type UpdatePhaseInput struct {
	Name            string
	ExpectedStartAt time.Time
	ActualStartAt   *time.Time
	ExpectedEndAt   time.Time
	ActualEndAt     *time.Time
	Status          entity.PhaseStatus
}

type Phase struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	cloudClientRegistry     *client.Registry
	authorizer              client.Authorizer
	featureToggles          feature.Toggles
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	projectDao              dao.Project
	phaseDao                dao.Phase
	storyDao                dao.Story
	projectPhaseRelationDao dao.ProjectPhaseRelation
	projectStoryRelationDao dao.ProjectStoryRelation
	phaseStoryRelationDao   dao.PhaseStoryRelation
}

func (p *Phase) FindPhases(ct context.Context, phaseFilter *PhaseFilter) ([]entity.Phase, *errs.Error) {
	var phases []entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			phases, err = p.phaseDao.FindPhasesWithTx(ct, tx)
			if err != nil {
				return err
			}

			if phaseFilter != nil {
				phases = filterPhases(phases, *phaseFilter)
			}

			return nil
		})

	return phases, transactionErr
}

func (p *Phase) FindStoriesByPhaseID(ct context.Context, phaseID uint64) ([]entity.Story, *errs.Error) {
	var stories []entity.Story
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			storyIDs, err := p.phaseStoryRelationDao.FindStoryIDsByPhaseIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			stories, err = p.storyDao.FindStoriesByIDsWithTx(ct, tx, storyIDs)
			return err
		})

	return stories, transactionErr
}

func (p *Phase) CreatePhase(ct context.Context, projectID uint64, input CreatePhaseInput) (entity.Phase, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Phase{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genPhaseIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "phaseID"}
	genPhaseIDRes, rpcErr := p.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genPhaseIDReq)
	if rpcErr != nil {
		return entity.Phase{}, errs.FromGRPCErr(rpcErr)
	}

	phase := entity.Phase{
		ID:              genPhaseIDRes.UniqueNumber,
		Name:            input.Name,
		Status:          entity.TodoPhaseStatus,
		ExpectedStartAt: input.ExpectedStartAt,
		ExpectedEndAt:   input.ExpectedEndAt,
		CreatorID:       userID,
		CreatedAt:       time.Now(),
	}

	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			err := p.phaseDao.CreatePhase(ct, tx, phase)
			if err != nil {
				return err
			}

			_, err = p.projectDao.FindProjectByIDWithTx(ct, tx, projectID)
			if err != nil {
				return err
			}

			projectPhaseRelation := entity.ProjectPhaseRelation{
				ProjectID: projectID,
				PhaseID:   phase.ID,
			}

			return p.projectPhaseRelationDao.CreateProjectPhaseRelation(ct, tx, projectPhaseRelation)
		})

	return phase, transactionErr
}

func (p *Phase) UpdatePhase(ct context.Context, phaseID uint64, input UpdatePhaseInput) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			now := time.Now()
			phase.Name = input.Name
			phase.ExpectedStartAt = input.ExpectedStartAt
			phase.ActualStartAt = input.ActualStartAt
			phase.ExpectedEndAt = input.ExpectedEndAt
			phase.ActualEndAt = input.ActualEndAt
			phase.Status = input.Status
			phase.UpdatedAt = &now

			return p.phaseDao.UpdatePhase(ct, tx, phase)
		})

	return phase, transactionErr
}

func (p *Phase) DeletePhase(ct context.Context, phaseID uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			err = p.projectPhaseRelationDao.DeleteProjectPhaseRelationsByPhaseID(ct, tx, phaseID)
			if err != nil {
				return err
			}

			err = p.phaseStoryRelationDao.DeletePhaseStoryRelationsByPhaseID(ct, tx, phaseID)
			if err != nil {
				return err
			}

			return p.phaseDao.DeletePhase(ct, tx, phaseID)
		})

	return phase, transactionErr
}

func (p *Phase) AddStoryToPhase(ct context.Context, phaseID uint64, storyID uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			story, err := p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
			if err != nil {
				return err
			}

			now := time.Now()
			story.IsPlanned = true
			story.UpdatedAt = &now

			updateStoryMutation := mutation.NewUpdateStory(
				p.logger,
				p.stateSyncer,
				p.storyDao,
				p.projectDao,
				p.projectStoryRelationDao,
				story,
			)

			rtTx.AppendMutation(updateStoryMutation)
			err = updateStoryMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			phaseStoryRelation := entity.PhaseStoryRelation{
				PhaseID: phaseID,
				StoryID: storyID,
			}

			return p.phaseStoryRelationDao.CreatePhaseStoryRelation(ct, tx, phaseStoryRelation)
		})

	return phase, transactionErr
}

func (p *Phase) AddStoriesToPhase(ct context.Context, phaseID uint64, storyIDs []uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			for _, storyID := range storyIDs {
				story, err := p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
				if err != nil {
					return err
				}

				now := time.Now()
				story.IsPlanned = true
				story.UpdatedAt = &now

				updateStoryMutation := mutation.NewUpdateStory(
					p.logger,
					p.stateSyncer,
					p.storyDao,
					p.projectDao,
					p.projectStoryRelationDao,
					story,
				)

				rtTx.AppendMutation(updateStoryMutation)
				err = updateStoryMutation.Execute(ct, tx)
				if err != nil {
					return err
				}

				phaseStoryRelation := entity.PhaseStoryRelation{
					PhaseID: phaseID,
					StoryID: storyID,
				}

				err = p.phaseStoryRelationDao.CreatePhaseStoryRelation(ct, tx, phaseStoryRelation)
				if err != nil {
					return err
				}
			}

			return nil
		})

	return phase, transactionErr
}

func (p *Phase) RemoveStoryFromPhase(ct context.Context, phaseID uint64, storyID uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			story, err := p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
			if err != nil {
				return err
			}

			now := time.Now()
			story.IsPlanned = false
			story.UpdatedAt = &now

			updateStoryMutation := mutation.NewUpdateStory(
				p.logger,
				p.stateSyncer,
				p.storyDao,
				p.projectDao,
				p.projectStoryRelationDao,
				story,
			)

			rtTx.AppendMutation(updateStoryMutation)
			err = updateStoryMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			return p.phaseStoryRelationDao.DeletePhaseStoryRelation(ct, tx, phaseID, storyID)
		})

	return phase, transactionErr
}

func (p *Phase) RemoveStoriesFromPhase(ct context.Context, phaseID uint64, storyIDs []uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	transactionErr := p.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
			if err != nil {
				return err
			}

			for _, storyID := range storyIDs {
				story, err := p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
				if err != nil {
					return err
				}

				now := time.Now()
				story.IsPlanned = false
				story.UpdatedAt = &now
				updateStoryMutation := mutation.NewUpdateStory(
					p.logger,
					p.stateSyncer,
					p.storyDao,
					p.projectDao,
					p.projectStoryRelationDao,
					story,
				)

				rtTx.AppendMutation(updateStoryMutation)
				err = updateStoryMutation.Execute(ct, tx)
				if err != nil {
					return err
				}

				return p.phaseStoryRelationDao.DeletePhaseStoryRelation(ct, tx, phaseID, storyID)
			}

			return nil
		})

	return phase, transactionErr
}

func NewPhase(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	projectDao dao.Project,
	phaseDao dao.Phase,
	storyDao dao.Story,
	projectPhaseRelationDao dao.ProjectPhaseRelation,
	projectStoryRelationDao dao.ProjectStoryRelation,
	phaseStoryRelationDao dao.PhaseStoryRelation,
) *Phase {
	return &Phase{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		cloudClientRegistry:     cloudClientRegistry,
		authorizer:              authorizer,
		featureToggles:          featureToggles,
		transactionFactory:      transactionFactory,
		stateSyncer:             stateSyncer,
		projectDao:              projectDao,
		phaseDao:                phaseDao,
		storyDao:                storyDao,
		projectPhaseRelationDao: projectPhaseRelationDao,
		projectStoryRelationDao: projectStoryRelationDao,
		phaseStoryRelationDao:   phaseStoryRelationDao,
	}
}
