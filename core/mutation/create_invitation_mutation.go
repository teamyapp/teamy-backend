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

type CreateInvitationMutation struct {
	logger        telemetry.Logger
	stateSyncer   *realtime.StateSyncer
	invitationDao dao.Invitation
	id            uint64
	invitation    entity.Invitation
}

var _ realtime.Mutation = (*CreateInvitationMutation)(nil)

func (c *CreateInvitationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateInvitationMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (c *CreateInvitationMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (c *CreateInvitationMutation) Execute(ct context.Context) *errs.Error {
	err := c.invitationDao.CreateInvitation(ct, c.invitation)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateInvitationMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateInvitationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.invitation.TeamID)
}

func (c *CreateInvitationMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	//TODO implement me
	panic("implement me")
}

func (c *CreateInvitationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.invitation,
	}
}

func (c *CreateInvitationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateInvitationMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitation entity.Invitation,
) *CreateInvitationMutation {
	return &CreateInvitationMutation{
		logger:        logger,
		stateSyncer:   stateSyncer,
		invitationDao: invitationDao,
		id:            stateSyncer.NextMutationID(),
		invitation:    invitation,
	}
}
