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

type Story struct {
	logger                  telemetry.Logger
	cloudClientRegistry     *client.Registry
	authorizer              client.Authorizer
	featureToggles          feature.Toggles
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	projectDao              dao.Project
	storyDao                dao.Story
	projectStoryRelationDao dao.ProjectStoryRelation
	storyTaskRelationDao    dao.StoryTaskRelation
	userDao                 dao.User
	taskDao                 dao.Task
}

func (s *Story) FindTasksByStoryID(ct context.Context, storyID uint64) ([]entity.Task, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	var tasks []entity.Task
	transactionErr := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		taskIDs, err := s.storyTaskRelationDao.FindTaskIDsByStoryIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		tasks, err = s.taskDao.FindTasksByIDsWithTx(ct, tx, taskIDs)
		return err
	})

	return tasks, transactionErr
}

func (s *Story) CreateStory(ct context.Context, projectID uint64, input CreateStoryInput) (entity.Story, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Story{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genStoryIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "storyID"}
	genStoryIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genStoryIDReq)
	if rpcErr != nil {
		return entity.Story{}, errs.FromGRPCErr(rpcErr)
	}

	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
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
		_, err := s.userDao.FindUserByIDWithTx(ct, tx, input.OwnerID)
		if err != nil {
			return err
		}

		err = s.storyDao.CreateStory(ct, tx, story)
		if err != nil {
			return err
		}

		_, err = s.projectDao.FindProjectByIDWithTx(ct, tx, projectID)
		if err != nil {
			return err
		}

		projectStoryRelation := entity.ProjectStoryRelation{
			ProjectID: projectID,
			StoryID:   story.ID,
		}

		return s.projectStoryRelationDao.CreateProjectStoryRelation(ct, tx, projectStoryRelation)
	})

	return story, transactionErr
}

func (s *Story) UpdateStory(ct context.Context, storyID uint64, input UpdateStoryInput) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		_, err = s.userDao.FindUserByIDWithTx(ct, tx, input.OwnerID)
		if err != nil {
			return err
		}

		story, err = s.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		now := time.Now()
		story.Name = input.Name
		story.OwnerID = input.OwnerID
		story.Status = input.Status
		story.Priority = input.Priority
		story.UpdatedAt = &now

		return s.storyDao.UpdateStory(ct, tx, story)
	})

	return story, transactionErr
}

func (s *Story) DeleteStory(ct context.Context, storyID uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = s.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		return s.storyDao.DeleteStory(ct, tx, storyID)
	})

	return story, transactionErr
}

func (s *Story) AddTaskToStory(ct context.Context, storyID uint64, taskID uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = s.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		_, err = s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		storyTaskRelation := entity.StoryTaskRelation{
			StoryID: storyID,
			TaskID:  taskID,
		}

		return s.storyTaskRelationDao.CreateStoryTaskRelation(ct, tx, storyTaskRelation)
	})

	return story, transactionErr
}

func (s *Story) AddTasksToStory(ct context.Context, storyID uint64, taskIDs []uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = s.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		for _, taskID := range taskIDs {
			_, err = s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
			if err != nil {
				return err
			}

			storyTaskRelation := entity.StoryTaskRelation{
				StoryID: storyID,
				TaskID:  taskID,
			}

			err = s.storyTaskRelationDao.CreateStoryTaskRelation(ct, tx, storyTaskRelation)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return story, transactionErr
}

func (s *Story) RemoveTaskFromStory(ct context.Context, storyID uint64, taskID uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		_, err = s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		story, err = s.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		return s.storyTaskRelationDao.DeleteStoryTaskRelation(ct, tx, storyID, taskID)
	})

	return story, transactionErr
}

func (s *Story) RemoveTasksFromStory(ct context.Context, storyID uint64, taskIDs []uint64) (entity.Story, *errs.Error) {
	var story entity.Story
	txCtx := transaction.NewTransactionsContext(
		s.logger,
		s.transactionFactory,
		s.stateSyncer,
		ct,
	)

	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		story, err = s.storyDao.FindStoryByIDWithTx(ct, tx, storyID)
		if err != nil {
			return err
		}

		for _, taskID := range taskIDs {
			_, err = s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
			if err != nil {
				return err
			}

			err = s.storyTaskRelationDao.DeleteStoryTaskRelation(ct, tx, storyID, taskID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return story, transactionErr
}

func NewStory(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	projectDao dao.Project,
	storyDao dao.Story,
	projectStoryRelationDao dao.ProjectStoryRelation,
	storyTaskRelationDao dao.StoryTaskRelation,
	userDao dao.User,
	taskDao dao.Task,
) *Story {
	return &Story{
		logger:                  logger,
		cloudClientRegistry:     cloudClientRegistry,
		authorizer:              authorizer,
		featureToggles:          featureToggles,
		transactionFactory:      transactionFactory,
		stateSyncer:             stateSyncer,
		projectDao:              projectDao,
		storyDao:                storyDao,
		projectStoryRelationDao: projectStoryRelationDao,
		storyTaskRelationDao:    storyTaskRelationDao,
		userDao:                 userDao,
		taskDao:                 taskDao,
	}
}
