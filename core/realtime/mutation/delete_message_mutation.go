package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteMessageMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	messageID     uint64
	messageDao    dao.Message
	dataCollector obs.DataCollector
}

func (c *DeleteMessageMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteMessageMutation) Execute(ct context.Context) error {
	err := d.messageDao.DeleteMessage(ct, d.messageID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteMessageMutation) Undo() error {
	return nil
}

func (d *DeleteMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.messageID,
	}
}

func NewDeleteMessageMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	messageID uint64,
	messageDao dao.Message,
	dataCollector obs.DataCollector) *DeleteMessageMutation {
	return &DeleteMessageMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		messageID:     messageID,
		messageDao:    messageDao,
		dataCollector: dataCollector,
	}
}
