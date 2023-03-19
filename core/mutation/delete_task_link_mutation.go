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

type DeleteTaskLinkMutation struct {
	dataCollector    telemetry.DataCollector
	stateSyncer      *realtime.StateSyncer
	taskLinkDaoV2    daov2.TaskLink
	taskDaoV2        daov2.Task
	id               uint64
	taskLink         entity.TaskLink
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTaskMutation)(nil)

func (d *DeleteTaskLinkMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTaskLinkMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := d.taskLinkDaoV2.DeleteTaskLink(ct, tx, d.taskLink.ID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteTaskLinkMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	task, err := d.taskDaoV2.FindTaskByID(ct, d.taskLink.TaskID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	var internalErr *errs.Error
	d.clientNotifiers, internalErr = d.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if internalErr != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteTaskLinkMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTaskLinkMutation) Execute(ct context.Context) *errs.Error {
	panic("deprecate me")
}

func (d *DeleteTaskLinkMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTaskLinkMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	panic("deprecate me")
}

func (d *DeleteTaskLinkMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskLinkCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.taskLink.ID,
	}
}

func (c *DeleteTaskLinkMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTaskLinkMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskLinkDaoV2 daov2.TaskLink,
	taskDaoV2 daov2.Task,
	taskLink entity.TaskLink,
) *DeleteTaskLinkMutation {
	return &DeleteTaskLinkMutation{
		dataCollector:    dataCollector,
		stateSyncer:      stateSyncer,
		taskLinkDaoV2:    taskLinkDaoV2,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		taskLink:         taskLink,
		notifierPrepared: false,
	}
}
