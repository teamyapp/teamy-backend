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

type CreateInvitation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	invitationDao     dao.Invitation
	id                uint64
	invitation        entity.Invitation
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateInvitation)(nil)

func (c *CreateInvitation) GetID() uint64 {
	return c.id
}

func (c *CreateInvitation) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.invitationDao.CreateInvitation(ct, tx, c.invitation)
}

func (c *CreateInvitation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.invitation.TeamID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateInvitation) Undo() *errs.Error {
	return nil
}

func (c *CreateInvitation) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateInvitation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.invitation,
	}
}

func (c *CreateInvitation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateInvitation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	invitation entity.Invitation,
) *CreateInvitation {
	return &CreateInvitation{
		logger:            logger,
		stateSyncer:       stateSyncer,
		invitationDao:     invitationDao,
		id:                stateSyncer.NextMutationID(),
		invitation:        invitation,
		notifiersPrepared: false,
	}
}
