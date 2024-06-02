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

type CreateTeamMemberGroupInvitationRelation struct {
	logger                               telemetry.Logger
	stateSyncer                          *realtime.StateSyncer
	teamMemberGroupInvitationRelationDao dao.TeamMemberGroupInvitationRelation
	teamMemberGroupDao                   dao.TeamMemberGroup
	teamDao                              dao.Team
	id                                   uint64
	teamMemberGroupInvitationRelation    entity.TeamMemberGroupInvitationRelation
	clientNotifiers                      []*realtime.ClientNotifier
	notifiersPrepared                    bool
}

var _ realtime.Mutation = (*CreateTeamMemberGroupInvitationRelation)(nil)

func (c *CreateTeamMemberGroupInvitationRelation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMemberGroupInvitationRelation) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.teamMemberGroupInvitationRelationDao.CreateTeamMemberGroupInvitationRelation(ct, tx, c.teamMemberGroupInvitationRelation)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateTeamMemberGroupInvitationRelation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	teamMemberGroup, err := c.teamMemberGroupDao.FindMemberGroupByID(ct, tx, c.teamMemberGroupInvitationRelation.GroupID)
	if err != nil {
		return err
	}

	team, err := c.teamDao.FindTeamByIDWithTx(ct, tx, teamMemberGroup.TeamID)
	if err != nil {
		return err
	}

	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, team.ID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateTeamMemberGroupInvitationRelation) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamMemberGroupInvitationRelation) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeamMemberGroupInvitationRelation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamMemberGroupInvitationRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.teamMemberGroupInvitationRelation,
	}
}

func (c *CreateTeamMemberGroupInvitationRelation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamMemberGroupInvitationRelation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamMemberGroupInvitationRelationDao dao.TeamMemberGroupInvitationRelation,
	teamMemberGroupDao dao.TeamMemberGroup,
	teamDao dao.Team,
	teamMemberGroupInvitationRelation entity.TeamMemberGroupInvitationRelation,
) *CreateTeamMemberGroupInvitationRelation {
	return &CreateTeamMemberGroupInvitationRelation{
		logger:                               logger,
		stateSyncer:                          stateSyncer,
		id:                                   stateSyncer.NextMutationID(),
		teamMemberGroupInvitationRelationDao: teamMemberGroupInvitationRelationDao,
		teamMemberGroupDao:                   teamMemberGroupDao,
		teamDao:                              teamDao,
		teamMemberGroupInvitationRelation:    teamMemberGroupInvitationRelation,
		notifiersPrepared:                    false,
	}
}
