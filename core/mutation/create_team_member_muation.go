package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTeamMemberMutation struct {
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	dataCollector obs.DataCollector
	id            uint64
	teamMember    entity.TeamMember
}

func (c *CreateTeamMemberMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMemberMutation) Execute(ct context.Context) error {
	err := c.teamMemberDao.CreateTeamMember(ct, c.teamMember)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	userNotifier, err := c.stateSyncer.GetUserNotifier(ct, c.teamMember.UserID)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	} else {
		err = c.stateSyncer.SubscribeToTeams(ct, c.teamMember.UserID, userNotifier)
		if err != nil {
			c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}
	}

	return nil
}

func (c *CreateTeamMemberMutation) Undo() error {
	return nil
}

func (c *CreateTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
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

func NewCreateTeamMemberMutation(
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	dataCollector obs.DataCollector,
	teamMember entity.TeamMember) *CreateTeamMemberMutation {
	return &CreateTeamMemberMutation{
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		teamMember:    teamMember,
	}
}
