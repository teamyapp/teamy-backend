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

type DeleteInvitation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	invitationDao     dao.Invitation
	invitationDaoV2   daov2.Invitation
	id                uint64
	invitation        entity.Invitation
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*DeleteInvitation)(nil)

func (d *DeleteInvitation) GetID() uint64 {
	return d.id
}

func (d *DeleteInvitation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.invitationDaoV2.DeleteInvitation(ct, tx, d.invitation.ID)
}

func (d *DeleteInvitation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	d.clientNotifiers, err = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.invitation.TeamID)
	if err != nil {
		return err
	}

	d.notifiersPrepared = true
	return nil
}

func (d *DeleteInvitation) Execute(ct context.Context) *errs.Error {
	err := d.invitationDao.DeleteInvitation(ct, d.invitation.ID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteInvitation) Undo() *errs.Error {
	return nil
}

func (d *DeleteInvitation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.invitation.TeamID)
}

func (d *DeleteInvitation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteInvitation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.invitation.ID,
	}
}

func (d *DeleteInvitation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteInvitation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitationDaoV2 daov2.Invitation,
	invitation entity.Invitation,
) *DeleteInvitation {
	return &DeleteInvitation{
		logger:          logger,
		stateSyncer:     stateSyncer,
		invitationDao:   invitationDao,
		invitationDaoV2: invitationDaoV2,
		id:              stateSyncer.NextMutationID(),
		invitation:      invitation,
	}
}
