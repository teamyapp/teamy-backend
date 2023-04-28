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

type CreateInvitation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	invitationDaoV2   daov2.Invitation
	id                uint64
	invitation        entity.Invitation
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateInvitation)(nil)

func (c *CreateInvitation) GetID() uint64 {
	return c.id
}

func (c *CreateInvitation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.invitationDaoV2.CreateInvitation(ct, tx, c.invitation)
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

func (c *CreateInvitation) GetClientNotifiersV2() []*realtime.ClientNotifier {
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
	invitationDaoV2 daov2.Invitation,
	invitation entity.Invitation,
) *CreateInvitation {
	return &CreateInvitation{
		logger:            logger,
		stateSyncer:       stateSyncer,
		invitationDaoV2:   invitationDaoV2,
		id:                stateSyncer.NextMutationID(),
		invitation:        invitation,
		notifiersPrepared: false,
	}
}
