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

type DeleteMessage struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	message          entity.Message
	messageDao       dao.Message
	messageDaoV2     daov2.Message
	id               uint64
	taskDao          dao.Task
	taskDaoV2        daov2.Task
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteMessage)(nil)

func (d *DeleteMessage) GetID() uint64 {
	return d.id
}

func (d *DeleteMessage) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.messageDaoV2.DeleteMessage(ct, tx, d.message.ID)
}

func (d *DeleteMessage) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	task, err := d.taskDaoV2.FindTaskByCommentsThreadIDWithTx(ct, tx, d.message.ThreadID)
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

func (d *DeleteMessage) Execute(ct context.Context) *errs.Error {
	err := d.messageDao.DeleteMessage(ct, d.message.ID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteMessage) Undo() *errs.Error {
	return nil
}

func (d *DeleteMessage) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	task, err := d.taskDao.FindTaskByCommentsThreadID(ct, d.message.ThreadID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return d.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (d *DeleteMessage) GetClientNotifiersV2() []*realtime.ClientNotifier {
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
	messageDaoV2 daov2.Message,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	message entity.Message,
) *DeleteMessage {
	return &DeleteMessage{
		logger:       logger,
		stateSyncer:  stateSyncer,
		messageDao:   messageDao,
		messageDaoV2: messageDaoV2,
		taskDao:      taskDao,
		taskDaoV2:    taskDaoV2,
		id:           stateSyncer.NextMutationID(),
		message:      message,
	}
}
