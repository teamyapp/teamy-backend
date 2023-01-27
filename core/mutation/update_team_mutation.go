package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMutation struct {
	dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamDao       dao.Team
	id            uint64
	team          entity.Team
}

func (u *UpdateTeamMutation) GetID() uint64 {
	return u.id
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

func (u *UpdateTeamMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewUpdateTeamMutation(
	dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	team entity.Team,
) *UpdateTeamMutation {
	return &UpdateTeamMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		id:            stateSyncer.NextMutationID(),
		team:          team,
	}
}
