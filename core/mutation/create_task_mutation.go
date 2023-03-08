package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskMutation struct {
	dataCollector    telemetry.DataCollector
	stateSyncer      *realtime.StateSyncer
	taskDao          dao.Task
	taskDaoV2        daov2.Task
	id               uint64
	task             entity.Task
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTaskMutation)(nil)

func (c *CreateTaskMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	internalErr := c.taskDaoV2.CreateTask(ct, tx, c.task)
	if internalErr != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (c *CreateTaskMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	if c.notifierPrepared {
		return nil
	}
	var internalErr *errs.Error
	c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.task.OwningTeamID)
	if internalErr != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateTaskMutation) Execute(ct context.Context) *errs.Error {
	err := c.taskDao.CreateTask(ct, c.task)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateTaskMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.task.OwningTeamID)
}

func (c *CreateTaskMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.task,
	}
}

func (c *CreateTaskMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTaskMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	task entity.Task,
) *CreateTaskMutation {
	return &CreateTaskMutation{
		dataCollector:    dataCollector,
		stateSyncer:      stateSyncer,
		taskDao:          taskDao,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		task:             task,
		notifierPrepared: false,
	}
}
