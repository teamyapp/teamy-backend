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

type UpdateTask struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	taskDao          dao.Task
	id               uint64
	task             entity.Task
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateTask)(nil)

func (u *UpdateTask) GetID() uint64 {
	return u.id
}

func (u *UpdateTask) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := u.taskDao.UpdateTask(ct, tx, u.task)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (u *UpdateTask) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	u.clientNotifiers, internalErr = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateTask) Undo() *errs.Error {
	return nil
}

func (u *UpdateTask) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateTask) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.task,
	}
}

func (u *UpdateTask) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTask(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	task entity.Task,
) *UpdateTask {
	return &UpdateTask{
		logger:           logger,
		stateSyncer:      stateSyncer,
		taskDao:          taskDao,
		id:               stateSyncer.NextMutationID(),
		task:             task,
		notifierPrepared: false,
	}
}
