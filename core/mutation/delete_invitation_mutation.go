package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteInvitationMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	id            uint64
	invitation    entity.Invitation
}

func (d *DeleteInvitationMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteInvitationMutation) Execute(ct context.Context) *errs.Error {
	err := d.invitationDao.DeleteInvitation(ct, d.invitation.ID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteInvitationMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteInvitationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
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

func (d *DeleteInvitationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteInvitationMutation(
	dataCollector telemetry.DataCollector,
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
