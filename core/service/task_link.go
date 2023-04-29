package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskLinkInput struct {
	TaskID       uint64
	Title        string
	URL          string
	IconURL      *string
	IconHoverURL *string
}

type TaskLink struct {
	logger              telemetry.Logger
	cloudClientRegistry *client.Registry
	authorizer          client.Authorizer
	featureToggles      feature.Toggles
	transactionFactory  transaction.Factory
	stateSyncer         *realtime.StateSyncer
	taskLinkDaoV2       daov2.TaskLink
	taskDaoV2           daov2.Task
}

func (t TaskLink) CreateTaskLink(ct context.Context, taskLinkEntity CreateTaskLinkInput) (entity.TaskLink, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.TaskLink{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewCreateLinkInTaskQuery(userID, taskLinkEntity.TaskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.TaskLink{}, err
		}

		if !hasPermission {
			return entity.TaskLink{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	genTaskLinkIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskLinkID"}
	genTaskLinkIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskLinkIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.TaskLink{}, internalErr
	}

	taskLink := entity.TaskLink{
		ID:           genTaskLinkIDRes.UniqueNumber,
		TaskID:       taskLinkEntity.TaskID,
		Title:        taskLinkEntity.Title,
		URL:          taskLinkEntity.URL,
		IconURL:      taskLinkEntity.IconURL,
		IconHoverURL: taskLinkEntity.IconHoverURL,
		CreatedAt:    time.Now(),
	}

	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}

	internalErr := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		createTaskLinkMutation := mutation.NewCreateTaskLink(t.logger, t.stateSyncer, t.taskLinkDaoV2, t.taskDaoV2, taskLink)
		internalErr := createTaskLinkMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTaskLinkMutation)
		return nil
	})

	if internalErr != nil {
		return entity.TaskLink{}, internalErr
	}

	if t.featureToggles.EnableAuthorization {
		err := t.authorizer.RegisterResource(ct, authorization.TaskLinkResourceType, taskLink.ID)
		if err != nil {
			return entity.TaskLink{}, err
		}

		err = t.authorizer.AssignParentResource(ct, authorization.TaskLinkResourceType, taskLink.ID, authorization.TaskResourceType, taskLink.TaskID)
		if err != nil {
			return entity.TaskLink{}, err
		}
	}

	return taskLink, nil
}

func (t TaskLink) DeleteTaskLink(ct context.Context, taskLinkID uint64) (entity.TaskLink, *errs.Error) {
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	var taskLink entity.TaskLink
	internalErr := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		taskLink, err = t.taskLinkDaoV2.FindTaskLinkByID(ct, tx, taskLinkID)
		if err != nil {
			return err
		}

		deleteTaskLinkMutation := mutation.NewDeleteTaskLink(t.logger, t.stateSyncer, t.taskLinkDaoV2, t.taskDaoV2, taskLink)
		internalErr := deleteTaskLinkMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(deleteTaskLinkMutation)
		return nil
	})

	if internalErr != nil {
		return entity.TaskLink{}, internalErr
	}

	return taskLink, nil
}

func (t TaskLink) FindLinksByTaskID(ct context.Context, taskID uint64) ([]entity.TaskLink, *errs.Error) {
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}

	var taskLinks []entity.TaskLink
	internalErr := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		taskLinks, err = t.taskLinkDaoV2.FindLinksByTaskID(ct, tx, taskID)
		return err
	})

	if internalErr != nil {
		return nil, internalErr
	}

	return taskLinks, nil
}

func NewTaskLink(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory transaction.Factory,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	taskLinkDaoV2 daov2.TaskLink,
	taskDaoV2 daov2.Task,
) TaskLink {
	return TaskLink{
		logger:              logger,
		cloudClientRegistry: cloudClientRegistry,
		transactionFactory:  transactionFactory,
		authorizer:          authorizer,
		featureToggles:      featureToggles,
		stateSyncer:         stateSyncer,
		taskLinkDaoV2:       taskLinkDaoV2,
		taskDaoV2:           taskDaoV2,
	}
}
