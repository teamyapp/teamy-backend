package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateInvitation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	invitationDaoV2   daov2.Invitation
	id                uint64
	invitation        entity.Invitation
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*UpdateInvitation)(nil)

func (u *UpdateInvitation) GetID() uint64 {
	return u.id
}

func (u *UpdateInvitation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.invitationDaoV2.UpdateInvitation(ct, tx, u.invitation)
}

func (u *UpdateInvitation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (u *UpdateInvitation) Undo() *errs.Error {
	return nil
}

func (u *UpdateInvitation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateInvitation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.invitation,
	}
}

func (u *UpdateInvitation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateInvitation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	invitationDaoV2 daov2.Invitation,
	invitation entity.Invitation,
) *UpdateInvitation {
	return &UpdateInvitation{
		logger:          logger,
		stateSyncer:     stateSyncer,
		invitationDaoV2: invitationDaoV2,
		id:              stateSyncer.NextMutationID(),
		invitation:      invitation,
	}
}
