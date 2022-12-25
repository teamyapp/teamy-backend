package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateInvitationMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	invitation    entity.Invitation
	invitationDao dao.Invitation
	dataCollector obs.DataCollector
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
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamID)
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
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	invitation entity.Invitation,
	invitationDao dao.Invitation,
	dataCollector obs.DataCollector) *UpdateInvitationMutation {
	return &UpdateInvitationMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		invitation:    invitation,
		invitationDao: invitationDao,
		dataCollector: dataCollector,
	}
}
