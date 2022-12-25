package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTeamMemberMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	teamMember    entity.TeamMember
	teamMemberDao dao.TeamMember
	dataCollector obs.DataCollector
}

func (c *CreateTeamMemberMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMemberMutation) Execute(ct context.Context) error {
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

	err = c.teamMemberDao.CreateTeamMember(ct, c.teamMember)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateTeamMemberMutation) Undo() error {
	return nil
}

func (c *CreateTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
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
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	teamMember entity.TeamMember,
	teamMemberDao dao.TeamMember,
	dataCollector obs.DataCollector) *CreateTeamMemberMutation {
	return &CreateTeamMemberMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		teamMember:    teamMember,
		teamMemberDao: teamMemberDao,
		dataCollector: dataCollector,
	}
}
