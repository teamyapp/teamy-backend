package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskAwaitForRelation struct {
	logger                  telemetry.Logger
	stateSyncer             *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskDao                 dao.Task
	id                      uint64
	taskAwaitForRelation    entity.TaskAwaitForRelation
	clientNotifiers         []*realtime.ClientNotifier
	notifiersPrepared       bool
}

var _ realtime.Mutation = (*CreateTaskAwaitForRelation)(nil)

func (c *CreateTaskAwaitForRelation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskAwaitForRelation) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.taskAwaitForRelationDao.CreateRelation(ct, tx, c.taskAwaitForRelation)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateTaskAwaitForRelation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	task, err := c.taskDao.FindTaskByIDWithTx(ct, tx, c.taskAwaitForRelation.AwaitForTaskID)
	if err != nil {
		return err
	}

	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateTaskAwaitForRelation) Undo() *errs.Error {
	return nil
}

func (c *CreateTaskAwaitForRelation) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTaskAwaitForRelation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.taskAwaitForRelation,
	}
}

func (c *CreateTaskAwaitForRelation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTaskAwaitForRelation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskDao dao.Task,
	taskAwaitForRelation entity.TaskAwaitForRelation,
) *CreateTaskAwaitForRelation {
	return &CreateTaskAwaitForRelation{
		logger:                  logger,
		stateSyncer:             stateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		taskDao:                 taskDao,
		id:                      stateSyncer.NextMutationID(),
		taskAwaitForRelation:    taskAwaitForRelation,
		notifiersPrepared:       false,
	}
}
