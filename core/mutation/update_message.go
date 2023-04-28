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

type UpdateMessage struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	messageDaoV2     daov2.Message
	taskDaoV2        daov2.Task
	id               uint64
	message          entity.Message
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateMessage)(nil)

func (u *UpdateMessage) GetID() uint64 {
	return u.id
}

func (u *UpdateMessage) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.messageDaoV2.UpdateMessage(ct, tx, u.message)
}

func (u *UpdateMessage) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	task, err := u.taskDaoV2.FindTaskByCommentsThreadIDWithTx(ct, tx, u.message.ThreadID)
	if err != nil {
		return err
	}

	var internalErr *errs.Error
	u.clientNotifiers, internalErr = u.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateMessage) Undo() *errs.Error {
	return nil
}

func (u *UpdateMessage) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateMessage) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.message,
	}
}

func (u *UpdateMessage) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateMessage(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	messageDaoV2 daov2.Message,
	taskDaoV2 daov2.Task,
	message entity.Message,
) *UpdateMessage {
	return &UpdateMessage{
		logger:           logger,
		stateSyncer:      stateSyncer,
		messageDaoV2:     messageDaoV2,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		message:          message,
		notifierPrepared: false,
	}
}
