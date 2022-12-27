package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMutation struct {
	stateSyncer   *realtime.StateSyncer
	teamDao       dao.Team
	dataCollector obs.DataCollector
	id            uint64
	team          entity.Team
}

func (c *UpdateTeamMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateTeamMutation) Execute(ct context.Context) error {
	err := u.teamDao.UpdateTeam(ct, u.team)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateTeamMutation) Undo() error {
	return nil
}

func (u *UpdateTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.team.ID)
}

func (u *UpdateTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.team,
	}
}

func NewUpdateTeamMutation(
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	dataCollector obs.DataCollector,
	team entity.Team) *UpdateTeamMutation {
	return &UpdateTeamMutation{
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		team:          team,
	}
}
