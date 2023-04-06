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

type CreateTeamMemberMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamMemberDao    dao.TeamMember
	teamMemberDaoV2  daov2.TeamMember
	id               uint64
	teamMember       entity.TeamMember
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTeamMemberMutation)(nil)

func (c *CreateTeamMemberMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMemberMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (c *CreateTeamMemberMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (c *CreateTeamMemberMutation) Execute(ct context.Context) *errs.Error {
	err := c.teamMemberDao.CreateTeamMember(ct, c.teamMember)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	}

	userNotifier, err := c.stateSyncer.GetUserNotifier(ct, c.teamMember.UserID)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	} else {
		err = c.stateSyncer.SubscribeToTeams(ct, c.teamMember.UserID, userNotifier)
		if err != nil {
			c.logger.ErrorWithContext(ct, err)
			return err
		}
	}

	return nil
}

func (c *CreateTeamMemberMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamMember.TeamID)
}

func (c *CreateTeamMemberMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeamMemberMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.teamMember,
	}
}

func (c *CreateTeamMemberMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamMemberMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	teamMember entity.TeamMember,
) *CreateTeamMemberMutation {
	return &CreateTeamMemberMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamMemberDao:    teamMemberDao,
		teamMemberDaoV2:  teamMemberDaoV2,
		id:               stateSyncer.NextMutationID(),
		teamMember:       teamMember,
		notifierPrepared: false,
	}
}
