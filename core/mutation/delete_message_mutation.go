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

type DeleteMessageMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	message       entity.Message
	messageDao    dao.Message
	id            uint64
	taskDao       dao.Task
}

var _ realtime.Mutation = (*DeleteMessageMutation)(nil)

func (d *DeleteMessageMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteMessageMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteMessageMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteMessageMutation) Execute(ct context.Context) *errs.Error {
	err := d.messageDao.DeleteMessage(ct, d.message.ID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
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
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return d.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
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
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	message entity.Message,
) *DeleteMessageMutation {
	return &DeleteMessageMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		messageDao:    messageDao,
		taskDao:       taskDao,
		id:            stateSyncer.NextMutationID(),
		message:       message,
	}
}
