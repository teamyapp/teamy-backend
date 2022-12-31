package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteInvitationMutation struct {
	dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	id            uint64
	invitation    entity.Invitation
}

func (c *DeleteInvitationMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteInvitationMutation) Execute(ct context.Context) error {
	err := d.invitationDao.DeleteInvitation(ct, d.invitation.ID)
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
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.invitation.TeamID)
}

func (d *DeleteInvitationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.invitation.ID,
	}
}

func (d *DeleteInvitationMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewDeleteInvitationMutation(
	dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitation entity.Invitation,
) *DeleteInvitationMutation {
	return &DeleteInvitationMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
