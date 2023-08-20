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

type CreateTask struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	taskDao          dao.Task
	id               uint64
	task             entity.Task
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTask)(nil)

func (c *CreateTask) GetID() uint64 {
	return c.id
}

func (c *CreateTask) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := c.taskDao.CreateTask(ct, tx, c.task)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (c *CreateTask) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateTask) Undo() *errs.Error {
	return nil
}

func (c *CreateTask) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTask) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.task,
	}
}

func (c *CreateTask) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTask(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	task entity.Task,
) *CreateTask {
	return &CreateTask{
		logger:           logger,
		stateSyncer:      stateSyncer,
		taskDao:          taskDao,
		id:               stateSyncer.NextMutationID(),
		task:             task,
		notifierPrepared: false,
	}
}
