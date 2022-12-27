package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTeamMutation struct {
	stateSyncer   *realtime.StateSyncer
	teamDao       dao.Team
	dataCollector obs.DataCollector
	id            uint64
	team          entity.Team
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
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.team.ID)
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
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	dataCollector obs.DataCollector,
	team entity.Team) *CreateTeamMutation {
	return &CreateTeamMutation{
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		team:          team,
	}
}
