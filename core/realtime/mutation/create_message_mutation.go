package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessageMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	message       entity.Message
	messageDao    dao.Message
	dataCollector obs.DataCollector
}

func (c *CreateMessageMutation) GetID() uint64 {
	return c.id
}

func (c *CreateMessageMutation) Execute(ct context.Context) error {
	err := c.messageDao.CreateMessage(ct, c.message)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateMessageMutation) Undo() error {
	return nil
}

func (c *CreateMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
}

func (c *CreateMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.message,
	}
}

func NewCreateMessageMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	message entity.Message,
	messageDao dao.Message,
	dataCollector obs.DataCollector) *CreateMessageMutation {
	return &CreateMessageMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		message:       message,
		messageDao:    messageDao,
		dataCollector: dataCollector,
	}
}
