package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateMessageMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	messageDao    dao.Message
	taskDao       dao.Task
	id            uint64
	message       entity.Message
}

var _ realtime.Mutation = (*UpdateMessageMutation)(nil)

func (u *UpdateMessageMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateMessageMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateMessageMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateMessageMutation) Execute(ct context.Context) *errs.Error {
	err := u.messageDao.UpdateMessage(ct, u.message)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
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
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return u.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
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
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	message entity.Message,
) *UpdateMessageMutation {
	return &UpdateMessageMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		messageDao:    messageDao,
		taskDao:       taskDao,
		id:            stateSyncer.NextMutationID(),
		message:       message,
	}
}
