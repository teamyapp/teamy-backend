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

type DeleteTaskMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	taskDao          dao.Task
	taskDaoV2        daov2.Task
	id               uint64
	task             entity.Task
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTaskMutation)(nil)

func (d *DeleteTaskMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTaskMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := d.taskDaoV2.DeleteTask(ct, tx, d.task.ID)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (d *DeleteTaskMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	d.clientNotifiers, internalErr = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteTaskMutation) Execute(ct context.Context) *errs.Error {
	err := d.taskDao.DeleteTask(ct, d.task.ID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteTaskMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.task.OwningTeamID)
}

func (d *DeleteTaskMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.task.ID,
	}
}

func (d *DeleteTaskMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTaskMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	task entity.Task,
) *DeleteTaskMutation {
	return &DeleteTaskMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		taskDao:          taskDao,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		task:             task,
		notifierPrepared: false,
	}
}
