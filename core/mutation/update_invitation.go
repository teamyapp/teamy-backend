package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateInvitation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	invitationDao     dao.Invitation
	id                uint64
	invitation        entity.Invitation
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*UpdateInvitation)(nil)

func (u *UpdateInvitation) GetID() uint64 {
	return u.id
}

func (u *UpdateInvitation) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.invitationDao.UpdateInvitation(ct, tx, u.invitation)
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

func (u *UpdateInvitation) GetClientNotifiers() []*realtime.ClientNotifier {
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
	invitationDao dao.Invitation,
	invitation entity.Invitation,
) *UpdateInvitation {
	return &UpdateInvitation{
		logger:        logger,
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
