package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateMessageMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	message       entity.Message
	messageDao    dao.Message
	dataCollector obs.DataCollector
}

func (c *UpdateMessageMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateMessageMutation) Execute(ct context.Context) error {
	err := u.messageDao.UpdateMessage(ct, u.message)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateMessageMutation) Undo() error {
	return nil
}

func (u *UpdateMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamID)
}

func (u *UpdateMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.message,
	}
}

func NewUpdateMessageMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	message entity.Message,
	messageDao dao.Message,
	dataCollector obs.DataCollector) *UpdateMessageMutation {
	return &UpdateMessageMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		message:       message,
		messageDao:    messageDao,
		dataCollector: dataCollector,
	}
}
