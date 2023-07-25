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

type DeleteMessage struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	message          entity.Message
	messageDao       dao.Message
	id               uint64
	taskDao          dao.Task
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteMessage)(nil)

func (d *DeleteMessage) GetID() uint64 {
	return d.id
}

func (d *DeleteMessage) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.messageDao.DeleteMessage(ct, tx, d.message.ID)
}

func (d *DeleteMessage) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	task, err := d.taskDao.FindTaskByCommentsThreadIDWithTx(ct, tx, d.message.ThreadID)
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

func (d *DeleteMessage) Undo() *errs.Error {
	return nil
}

func (d *DeleteMessage) GetClientNotifiers() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteMessage) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.message.ID,
	}
}

func (d *DeleteMessage) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteMessage(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	message entity.Message,
) *DeleteMessage {
	return &DeleteMessage{
		logger:      logger,
		stateSyncer: stateSyncer,
		messageDao:  messageDao,
		taskDao:     taskDao,
		id:          stateSyncer.NextMutationID(),
		message:     message,
	}
}
