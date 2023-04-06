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

type DeleteInvitationMutation struct {
	logger        telemetry.Logger
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	id            uint64
	invitation    entity.Invitation
}

var _ realtime.Mutation = (*DeleteInvitationMutation)(nil)

func (d *DeleteInvitationMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteInvitationMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteInvitationMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteInvitationMutation) Execute(ct context.Context) *errs.Error {
	err := d.invitationDao.DeleteInvitation(ct, d.invitation.ID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
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

func (d *DeleteInvitationMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	//TODO implement me
	panic("implement me")
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
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitation entity.Invitation,
) *DeleteInvitationMutation {
	return &DeleteInvitationMutation{
		logger:        logger,
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
