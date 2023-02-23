package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTeamMemberMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	id            uint64
	teamMember    entity.TeamMember
}

var _ realtime.Mutation = (*CreateTeamMemberMutation)(nil)

func (c *CreateTeamMemberMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (c *CreateTeamMemberMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) ([]*realtime.ClientNotifier, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *CreateTeamMemberMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMemberMutation) Execute(ct context.Context) *errs.Error {
	err := c.teamMemberDao.CreateTeamMember(ct, c.teamMember)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	userNotifier, err := c.stateSyncer.GetUserNotifier(ct, c.teamMember.UserID)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	} else {
		err = c.stateSyncer.SubscribeToTeams(ct, c.teamMember.UserID, userNotifier)
		if err != nil {
			c.dataCollector.Logger.ErrorWithContext(ct, err)
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
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	teamMember entity.TeamMember,
) *CreateTeamMemberMutation {
	return &CreateTeamMemberMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		id:            stateSyncer.NextMutationID(),
		teamMember:    teamMember,
	}
}
