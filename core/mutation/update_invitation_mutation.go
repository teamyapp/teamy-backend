package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateInvitationMutation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	invitationDao     dao.Invitation
	invitationDaoV2   daov2.Invitation
	id                uint64
	invitation        entity.Invitation
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*UpdateInvitationMutation)(nil)

func (u *UpdateInvitationMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateInvitationMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.invitationDaoV2.UpdateInvitation(ct, tx, u.invitation)
}

func (u *UpdateInvitationMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.invitation.TeamID)
	if err != nil {
		return err
	}

	u.notifiersPrepared = true
	return nil
}

func (u *UpdateInvitationMutation) Execute(ct context.Context) *errs.Error {
	err := u.invitationDao.UpdateInvitation(ct, u.invitation)
	if err != nil {
		u.logger.ErrorWithContext(ct, err)
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

func (u *UpdateInvitationMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
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
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitationDaoV2 daov2.Invitation,
	invitation entity.Invitation,
) *UpdateInvitationMutation {
	return &UpdateInvitationMutation{
		logger:          logger,
		stateSyncer:     stateSyncer,
		invitationDao:   invitationDao,
		invitationDaoV2: invitationDaoV2,
		id:              stateSyncer.NextMutationID(),
		invitation:      invitation,
	}
}
