package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateInvitationMutation struct {
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	dataCollector obs.DataCollector
	id            uint64
	invitation    entity.Invitation
}

func (c *UpdateInvitationMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateInvitationMutation) Execute(ct context.Context) error {
	err := u.invitationDao.UpdateInvitation(ct, u.invitation)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateInvitationMutation) Undo() error {
	return nil
}

func (u *UpdateInvitationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.invitation.TeamID)
}

func (u *UpdateInvitationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.invitation,
	}
}

func NewUpdateInvitationMutation(
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	dataCollector obs.DataCollector,
	invitation entity.Invitation) *UpdateInvitationMutation {
	return &UpdateInvitationMutation{
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
