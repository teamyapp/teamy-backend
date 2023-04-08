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

type UpdateMessageMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	messageDao       dao.Message
	messageDaoV2     daov2.Message
	taskDao          dao.Task
	taskDaoV2        daov2.Task
	id               uint64
	message          entity.Message
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateMessageMutation)(nil)

func (u *UpdateMessageMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateMessageMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.messageDaoV2.UpdateMessage(ct, tx, u.message)
}

func (u *UpdateMessageMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (u *UpdateMessageMutation) Execute(ct context.Context) *errs.Error {
	err := u.messageDao.UpdateMessage(ct, u.message)
	if err != nil {
		u.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateMessageMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	task, err := u.taskDao.FindTaskByCommentsThreadID(ct, u.message.ThreadID)
	if err != nil {
		u.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return u.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (u *UpdateMessageMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.message,
	}
}

func (u *UpdateMessageMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateMessageMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	messageDaoV2 daov2.Message,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	message entity.Message,
) *UpdateMessageMutation {
	return &UpdateMessageMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		messageDao:       messageDao,
		messageDaoV2:     messageDaoV2,
		taskDao:          taskDao,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		message:          message,
		notifierPrepared: false,
	}
}
