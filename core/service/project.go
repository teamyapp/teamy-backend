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
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type Project struct {
	logger                  telemetry.Logger
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
	storyTaskRelationDao    dao.StoryTaskRelation
	userDao                 dao.User
	taskDao                 dao.Task
}

type CreateProjectInput struct {
	Name            string
	ExpectedStartAt *time.Time
	ExpectedEndAt   *time.Time
}

type UpdateProjectInput struct {
	Name            string
	ExpectedStartAt *time.Time
	ActualStartAt   *time.Time
	ExpectedEndAt   *time.Time
	ActualEndAt     *time.Time
}

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

type CreateStoryInput struct {
	Name     string
	OwnerID  uint64
	Priority entity.Priority
}

type UpdateStoryInput struct {
	Name     string
	OwnerID  uint64
	Status   entity.StoryStatus
	Priority entity.Priority
}

func (p *Project) FindStoriesByProjectID(ct context.Context, projectID uint64) ([]entity.Story, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	var stories []entity.Story
	transactionErr := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		storyIDs, err := p.projectStoryRelationDao.FindStoryIDsByProjectIDWithTx(ct, tx, projectID)
		if err != nil {
			return err
		}

		stories, err = p.storyDao.FindStoriesByIDsWithTx(ct, tx, storyIDs)
		return err
	})

	return stories, transactionErr
}

func (p *Project) FindStoriesByPhaseID(ct context.Context, phaseID uint64) ([]entity.Story, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	var stories []entity.Story
	transactionErr := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		storyIDs, err := p.phaseStoryRelationDao.FindStoryIDsByPhaseIDWithTx(ct, tx, phaseID)
		if err != nil {
			return err
		}

		stories, err = p.storyDao.FindStoriesByIDsWithTx(ct, tx, storyIDs)
		return err
	})

	return stories, transactionErr
}

func (p *Project) FindPhasesByProjectID(ct context.Context, projectID uint64) ([]entity.Phase, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	var phases []entity.Phase
	transactionErr := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		phaseIDs, err := p.projectPhaseRelationDao.FindPhaseIDsByProjectIDWithTx(ct, tx, projectID)
		if err != nil {
			return err
		}

		phases, err = p.phaseDao.FindPhasesByIDsWithTx(ct, tx, phaseIDs)
		return err
	})

	return phases, transactionErr
}

func (p *Project) FindTasksByStoryID(ct context.Context, storyID uint64) ([]entity.Task, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	var tasks []entity.Task
	transactionErr := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		taskIDs, err := p.storyTaskRelationDao.FindTaskIDsByStoryIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		tasks, err = p.taskDao.FindTasksByIDsWithTx(ct, tx, taskIDs)
		return err
	})

	return tasks, transactionErr
}

func (p *Project) CreateProject(ct context.Context, input CreateProjectInput) (entity.Project, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Project{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genProjectIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "projectID"}
	genProjectIDRes, rpcErr := p.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genProjectIDReq)
	if rpcErr != nil {
		return entity.Project{}, errs.FromGRPCErr(rpcErr)
	}

	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	project := entity.Project{
		ID:              genProjectIDRes.UniqueNumber,
		Name:            input.Name,
		ExpectedStartAt: input.ExpectedStartAt,
		ExpectedEndAt:   input.ExpectedEndAt,
		CreatorID:       userID,
		CreatedAt:       time.Now(),
	}

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return p.projectDao.CreateProject(ct, tx, project)
	})

	return project, transactionErr
}

func (p *Project) UpdateProject(ct context.Context, projectID uint64, input UpdateProjectInput) (entity.Project, *errs.Error) {
	var project entity.Project
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

		return p.projectDao.UpdateProject(ct, tx, project)
	})

	return project, transactionErr
}

func (p *Project) DeleteProject(ct context.Context, projectID uint64) (entity.Project, *errs.Error) {
	var project entity.Project
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		project, err = p.projectDao.FindProjectByIDWithTx(ct, tx, projectID)
		if err != nil {
			return err
		}

		return p.projectDao.DeleteProject(ct, tx, projectID)
	})

	return project, transactionErr
}

func (p *Project) CreatePhase(ct context.Context, projectID uint64, input CreatePhaseInput) (entity.Phase, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Phase{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genPhaseIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "phaseID"}
	genPhaseIDRes, rpcErr := p.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genPhaseIDReq)
	if rpcErr != nil {
		return entity.Phase{}, errs.FromGRPCErr(rpcErr)
	}

	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	phase := entity.Phase{
		ID:              genPhaseIDRes.UniqueNumber,
		Name:            input.Name,
		Status:          entity.TodoPhaseStatus,
		ExpectedStartAt: input.ExpectedStartAt,
		ExpectedEndAt:   input.ExpectedEndAt,
		CreatorID:       userID,
		CreatedAt:       time.Now(),
	}

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

func (p *Project) UpdatePhase(ct context.Context, phaseID uint64, input UpdatePhaseInput) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

func (p *Project) DeletePhase(ct context.Context, phaseID uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
		if err != nil {
			return err
		}

		return p.phaseDao.DeletePhase(ct, tx, phaseID)
	})

	return phase, transactionErr
}

func (p *Project) AddStoryToPhase(ct context.Context, phaseID uint64, storyID uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
		if err != nil {
			return err
		}

		_, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
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

func (p *Project) AddStoriesToPhase(ct context.Context, phaseID uint64, storyIDs []uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
		if err != nil {
			return err
		}

		for _, storyID := range storyIDs {
			_, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
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

func (p *Project) RemoveStoryFromPhase(ct context.Context, phaseID uint64, storyID uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		_, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
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

func (p *Project) RemoveStoriesFromPhase(ct context.Context, phaseID uint64, storyIDs []uint64) (entity.Phase, *errs.Error) {
	var phase entity.Phase
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		phase, err = p.phaseDao.FindPhaseByIDWithTx(ct, tx, phaseID)
		if err != nil {
			return err
		}

		for _, storyID := range storyIDs {
			_, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
			if err != nil {
				return err
			}

			err = p.phaseStoryRelationDao.DeletePhaseStoryRelation(ct, tx, phaseID, storyID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return phase, transactionErr
}

func (p *Project) CreateStory(ct context.Context, projectID uint64, input CreateStoryInput) (entity.Story, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Story{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genStoryIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "storyID"}
	genStoryIDRes, rpcErr := p.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genStoryIDReq)
	if rpcErr != nil {
		return entity.Story{}, errs.FromGRPCErr(rpcErr)
	}

	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	story := entity.Story{
		ID:        genStoryIDRes.UniqueNumber,
		Name:      input.Name,
		OwnerID:   input.OwnerID,
		Priority:  input.Priority,
		Status:    entity.TodoStoryStatus,
		CreatorID: userID,
		CreatedAt: time.Now(),
	}

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		_, err := p.userDao.FindUserByIDWithTx(ct, tx, input.OwnerID)
		if err != nil {
			return err
		}

		err = p.storyDao.CreateStory(ct, tx, story)
		if err != nil {
			return err
		}

		_, err = p.projectDao.FindProjectByIDWithTx(ct, tx, projectID)
		if err != nil {
			return err
		}

		projectStoryRelation := entity.ProjectStoryRelation{
			ProjectID: projectID,
			StoryID:   story.ID,
		}

		return p.projectStoryRelationDao.CreateProjectStoryRelation(ct, tx, projectStoryRelation)
	})

	return story, transactionErr
}

func (p *Project) UpdateStory(ct context.Context, storyID uint64, input UpdateStoryInput) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		_, err = p.userDao.FindUserByIDWithTx(ct, tx, input.OwnerID)
		if err != nil {
			return err
		}

		story, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		now := time.Now()
		story.Name = input.Name
		story.OwnerID = input.OwnerID
		story.Status = input.Status
		story.Priority = input.Priority
		story.UpdatedAt = &now

		return p.storyDao.UpdateStory(ct, tx, story)
	})

	return story, transactionErr
}

func (p *Project) DeleteStory(ct context.Context, storyID uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		return p.storyDao.DeleteStory(ct, tx, storyID)
	})

	return story, transactionErr
}

func (p *Project) AddTaskToStory(ct context.Context, storyID uint64, taskID uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		_, err = p.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		storyTaskRelation := entity.StoryTaskRelation{
			StoryID: storyID,
			TaskID:  taskID,
		}

		return p.storyTaskRelationDao.CreateStoryTaskRelation(ct, tx, storyTaskRelation)
	})

	return story, transactionErr
}

func (p *Project) AddTasksToStory(ct context.Context, storyID uint64, taskIDs []uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		for _, taskID := range taskIDs {
			_, err = p.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
			if err != nil {
				return err
			}

			storyTaskRelation := entity.StoryTaskRelation{
				StoryID: storyID,
				TaskID:  taskID,
			}

			err = p.storyTaskRelationDao.CreateStoryTaskRelation(ct, tx, storyTaskRelation)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return story, transactionErr
}

func (p *Project) RemoveTaskFromStory(ct context.Context, storyID uint64, taskID uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		_, err = p.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		story, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		return p.storyTaskRelationDao.DeleteStoryTaskRelation(ct, tx, storyID, taskID)
	})

	return story, transactionErr
}

func (p *Project) RemoveTasksFromStory(ct context.Context, storyID uint64, taskIDs []uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = p.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		for _, taskID := range taskIDs {
			_, err = p.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
			if err != nil {
				return err
			}

			err = p.storyTaskRelationDao.DeleteStoryTaskRelation(ct, tx, storyID, taskID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return story, transactionErr
}

func NewProject(
	logger telemetry.Logger,
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
	storyTaskRelationDao dao.StoryTaskRelation,
	userDao dao.User,
	taskDao dao.Task,
) *Project {
	return &Project{
		logger:                  logger,
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
		storyTaskRelationDao:    storyTaskRelationDao,
		userDao:                 userDao,
		taskDao:                 taskDao,
	}
}
