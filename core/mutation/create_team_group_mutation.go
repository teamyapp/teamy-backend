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

type CreateTeamGroupMutation struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	teamGroupDaoV2    daov2.TeamGroup
	id                uint64
	teamGroup         entity.TeamGroup
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateTeamGroupMutation)(nil)

func (c *CreateTeamGroupMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamGroupMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.teamGroupDaoV2.CreateGroup(ct, tx, c.teamGroup)
}

func (c *CreateTeamGroupMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamGroup.TeamID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateTeamGroupMutation) Execute(ct context.Context) *errs.Error {
	return nil
}

func (c *CreateTeamGroupMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamGroupMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamGroup.TeamID)
}

func (c *CreateTeamGroupMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeamGroupMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamGroupCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.teamGroup,
	}
}

func (c *CreateTeamGroupMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamGroupMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamGroupDaoV2 daov2.TeamGroup,
	teamGroup entity.TeamGroup,
) *CreateTeamGroupMutation {
	return &CreateTeamGroupMutation{
		logger:            logger,
		stateSyncer:       stateSyncer,
		teamGroupDaoV2:    teamGroupDaoV2,
		id:                stateSyncer.NextMutationID(),
		teamGroup:         teamGroup,
		notifiersPrepared: false,
	}
}
