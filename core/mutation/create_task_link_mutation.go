package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskLinkMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	taskLinkDaoV2    daov2.TaskLink
	taskDaoV2        daov2.Task
	id               uint64
	taskLink         entity.TaskLink
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTaskLinkMutation)(nil)

func (c *CreateTaskLinkMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskLinkMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.taskLinkDaoV2.CreateTaskLink(ct, tx, c.taskLink)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateTaskLinkMutation) Execute(ct context.Context) *errs.Error {
	panic("deprecate me")
}

func (c *CreateTaskLinkMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	panic("deprecate me")
}

func (c *CreateTaskLinkMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	task, err := c.taskDaoV2.FindTaskByIDWithTx(ct, tx, c.taskLink.TaskID)
	if err != nil {
		return err
	}

	var internalErr *errs.Error
	c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateTaskLinkMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTaskLinkMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTaskLinkMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskLinkCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.taskLink,
	}
}

func (c *CreateTaskLinkMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTaskLinkMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskLinkDaoV2 daov2.TaskLink,
	taskDaoV2 daov2.Task,
	taskLink entity.TaskLink,
) *CreateTaskLinkMutation {
	return &CreateTaskLinkMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		taskLinkDaoV2:    taskLinkDaoV2,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		taskLink:         taskLink,
		notifierPrepared: false,
	}
}
