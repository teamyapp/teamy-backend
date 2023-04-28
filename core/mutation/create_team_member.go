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

type CreateTeamMember struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamMemberDaoV2  daov2.TeamMember
	id               uint64
	teamMember       entity.TeamMember
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTeamMember)(nil)

func (c *CreateTeamMember) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMember) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.teamMemberDaoV2.CreateTeamMember(ct, tx, c.teamMember)
	if err != nil {
		return err
	}

	teamIDs, err := c.teamMemberDaoV2.FindTeamIDsByUserIDWithTx(ct, tx, c.teamMember.UserID)
	if err != nil {
		return err
	}

	userNotifier, err := c.stateSyncer.GetUserNotifierV2(ct, c.teamMember.UserID, teamIDs)
	if err != nil {
		return err
	}

	return c.stateSyncer.SubscribeToTeamsV2(ct, c.teamMember.UserID, userNotifier, teamIDs)
}

func (c *CreateTeamMember) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	var err *errs.Error
	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamMember.TeamID)
	if err != nil {
		return err
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateTeamMember) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamMember) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeamMember) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.teamMember,
	}
}

func (c *CreateTeamMember) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamMember(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamMemberDaoV2 daov2.TeamMember,
	teamMember entity.TeamMember,
) *CreateTeamMember {
	return &CreateTeamMember{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamMemberDaoV2:  teamMemberDaoV2,
		id:               stateSyncer.NextMutationID(),
		teamMember:       teamMember,
		notifierPrepared: false,
	}
}
