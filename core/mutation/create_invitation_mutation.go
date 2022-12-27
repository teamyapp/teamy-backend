package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateInvitationMutation struct {
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	dataCollector obs.DataCollector
	id            uint64
	invitation    entity.Invitation
}

func (c *CreateInvitationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateInvitationMutation) Execute(ct context.Context) error {
	err := c.invitationDao.CreateInvitation(ct, c.invitation)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateInvitationMutation) Undo() error {
	return nil
}

func (c *CreateInvitationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.invitation.TeamID)
}

func (c *CreateInvitationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.invitation,
	}
}

func NewCreateInvitationMutation(
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	dataCollector obs.DataCollector,
	invitation entity.Invitation) *CreateInvitationMutation {
	return &CreateInvitationMutation{
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
