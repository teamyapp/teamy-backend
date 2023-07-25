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

type CreateTeamMember struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamMemberDao    dao.TeamMember
	id               uint64
	teamMember       entity.TeamMember
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTeamMember)(nil)

func (c *CreateTeamMember) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMember) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.teamMemberDao.CreateTeamMember(ct, tx, c.teamMember)
	if err != nil {
		return err
	}

	teamIDs, err := c.teamMemberDao.FindTeamIDsByUserIDWithTx(ct, tx, c.teamMember.UserID)
	if err != nil {
		return err
	}

	userNotifier, err := c.stateSyncer.GetUserNotifier(ct, c.teamMember.UserID, teamIDs)
	if err != nil {
		return err
	}

	return c.stateSyncer.SubscribeToTeams(ct, c.teamMember.UserID, userNotifier, teamIDs)
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

func (c *CreateTeamMember) GetClientNotifiers() []*realtime.ClientNotifier {
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
	teamMemberDao dao.TeamMember,
	teamMember entity.TeamMember,
) *CreateTeamMember {
	return &CreateTeamMember{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamMemberDao:    teamMemberDao,
		id:               stateSyncer.NextMutationID(),
		teamMember:       teamMember,
		notifierPrepared: false,
	}
}
