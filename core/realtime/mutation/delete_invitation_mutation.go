package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteInvitationMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	invitationID  uint64
	invitationDao dao.Invitation
	dataCollector obs.DataCollector
}

func (c *DeleteInvitationMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteInvitationMutation) Execute(ct context.Context) error {
	err := d.invitationDao.DeleteInvitation(ct, d.invitationID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteInvitationMutation) Undo() error {
	return nil
}

func (d *DeleteInvitationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteInvitationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.invitationID,
	}
}

func NewDeleteInvitationMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	invitationID uint64,
	invitationDao dao.Invitation,
	dataCollector obs.DataCollector) *DeleteInvitationMutation {
	return &DeleteInvitationMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		invitationID:  invitationID,
		invitationDao: invitationDao,
		dataCollector: dataCollector,
	}
}
