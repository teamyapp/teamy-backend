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

type DeleteTaskLink struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	taskLinkDao      dao.TaskLink
	taskDao          dao.Task
	id               uint64
	taskLink         entity.TaskLink
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTask)(nil)

func (d *DeleteTaskLink) GetID() uint64 {
	return d.id
}

func (d *DeleteTaskLink) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := d.taskLinkDao.DeleteTaskLink(ct, tx, d.taskLink.ID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeleteTaskLink) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	task, err := d.taskDao.FindTaskByIDWithTx(ct, tx, d.taskLink.TaskID)
	if err != nil {
		return err
	}

	var internalErr *errs.Error
	d.clientNotifiers, internalErr = d.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteTaskLink) GetClientNotifiers() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTaskLink) Undo() *errs.Error {
	return nil
}

func (d *DeleteTaskLink) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskLinkCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.taskLink.ID,
	}
}

func (c *DeleteTaskLink) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTaskLink(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskLinkDao dao.TaskLink,
	taskDao dao.Task,
	taskLink entity.TaskLink,
) *DeleteTaskLink {
	return &DeleteTaskLink{
		logger:           logger,
		stateSyncer:      stateSyncer,
		taskLinkDao:      taskLinkDao,
		taskDao:          taskDao,
		id:               stateSyncer.NextMutationID(),
		taskLink:         taskLink,
		notifierPrepared: false,
	}
}
