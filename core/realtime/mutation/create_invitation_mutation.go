package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateInvitationMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	invitation    entity.Invitation
	invitationDao dao.Invitation
	dataCollector obs.DataCollector
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
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
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
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	invitation entity.Invitation,
	invitationDao dao.Invitation,
	dataCollector obs.DataCollector) *CreateInvitationMutation {
	return &CreateInvitationMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		invitation:    invitation,
		invitationDao: invitationDao,
		dataCollector: dataCollector,
	}
}
