package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskAwaitForRelationMutation struct {
	dataCollector             telemetry.DataCollector
	stateSyncer               *realtime.StateSyncer
	taskAwaitForRelationDao   dao.TaskAwaitForRelation
	taskAwaitForRelationDaoV2 daov2.TaskAwaitForRelation
	taskDao                   dao.Task
	taskDaoV2                 daov2.Task
	id                        uint64
	taskAwaitForRelation      entity.TaskAwaitForRelation
	clientNotifiers           []*realtime.ClientNotifier
	notifiersPrepared         bool
}

var _ realtime.Mutation = (*CreateTaskAwaitForRelationMutation)(nil)

func (c *CreateTaskAwaitForRelationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskAwaitForRelationMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.taskAwaitForRelationDaoV2.CreateRelation(ct, tx, c.taskAwaitForRelation)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateTaskAwaitForRelationMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	task, err := c.taskDaoV2.FindTaskByIDWithTx(ct, tx, c.taskAwaitForRelation.AwaitForTaskID)
	if err != nil {
		return err
	}

	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateTaskAwaitForRelationMutation) Execute(ct context.Context) *errs.Error {
	err := c.taskAwaitForRelationDao.CreateRelation(ct, c.taskAwaitForRelation)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateTaskAwaitForRelationMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTaskAwaitForRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	task, err := c.taskDao.FindTaskByID(ct, c.taskAwaitForRelation.AwaitForTaskID)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (c *CreateTaskAwaitForRelationMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTaskAwaitForRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.taskAwaitForRelation,
	}
}

func (c *CreateTaskAwaitForRelationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTaskAwaitForRelationMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskAwaitForRelationDaoV2 daov2.TaskAwaitForRelation,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	taskAwaitForRelation entity.TaskAwaitForRelation,
) *CreateTaskAwaitForRelationMutation {
	return &CreateTaskAwaitForRelationMutation{
		dataCollector:             dataCollector,
		stateSyncer:               stateSyncer,
		taskAwaitForRelationDao:   taskAwaitForRelationDao,
		taskAwaitForRelationDaoV2: taskAwaitForRelationDaoV2,
		taskDao:                   taskDao,
		taskDaoV2:                 taskDaoV2,
		id:                        stateSyncer.NextMutationID(),
		taskAwaitForRelation:      taskAwaitForRelation,
		notifiersPrepared:         false,
	}
}
