package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateInvitationMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	id            uint64
	invitation    entity.Invitation
}

var _ realtime.Mutation = (*UpdateInvitationMutation)(nil)

func (u *UpdateInvitationMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateInvitationMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) ([]*realtime.ClientNotifier, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateInvitationMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateInvitationMutation) Execute(ct context.Context) *errs.Error {
	err := u.invitationDao.UpdateInvitation(ct, u.invitation)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateInvitationMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateInvitationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
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

func (u *UpdateInvitationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateInvitationMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitation entity.Invitation,
) *UpdateInvitationMutation {
	return &UpdateInvitationMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
