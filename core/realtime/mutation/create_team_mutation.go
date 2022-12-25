package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTeamMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	team          entity.Team
	teamDao       dao.Team
	dataCollector obs.DataCollector
}

func (c *CreateTeamMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMutation) Execute(ct context.Context) error {
	err := c.teamDao.CreateTeam(ct, c.team)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateTeamMutation) Undo() error {
	return nil
}

func (c *CreateTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
}

func (c *CreateTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.team,
	}
}

func NewCreateTeamMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	team entity.Team,
	teamDao dao.Team,
	dataCollector obs.DataCollector) *CreateTeamMutation {
	return &CreateTeamMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		team:          team,
		teamDao:       teamDao,
		dataCollector: dataCollector,
	}
}
