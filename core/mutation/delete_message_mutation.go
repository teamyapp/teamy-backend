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

type DeleteMessageMutation struct {
	logger      telemetry.Logger
	stateSyncer *realtime.StateSyncer
	message     entity.Message
	messageDao  dao.Message
	id          uint64
	taskDao     dao.Task
}

var _ realtime.Mutation = (*DeleteMessageMutation)(nil)

func (d *DeleteMessageMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteMessageMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteMessageMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteMessageMutation) Execute(ct context.Context) *errs.Error {
	err := d.messageDao.DeleteMessage(ct, d.message.ID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteMessageMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	task, err := d.taskDao.FindTaskByCommentsThreadID(ct, d.message.ThreadID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return d.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (d *DeleteMessageMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.message.ID,
	}
}

func (d *DeleteMessageMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteMessageMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	message entity.Message,
) *DeleteMessageMutation {
	return &DeleteMessageMutation{
		logger:      logger,
		stateSyncer: stateSyncer,
		messageDao:  messageDao,
		taskDao:     taskDao,
		id:          stateSyncer.NextMutationID(),
		message:     message,
	}
}
