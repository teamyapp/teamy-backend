package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TeamSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	teamDao             dao.Team
}

func (t TeamSyncer) CreateAndSyncTeam(ct context.Context, team entity.Team) error {
	err := t.teamDao.CreateTeam(ct, team)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			team.ID,
		},
		Payload: team,
	})
	return nil
}

func (t TeamSyncer) UpdateAndSyncTeam(ct context.Context, team entity.Team) error {
	err := t.teamDao.UpdateTeam(ct, team)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs: []uint64{
			team.ID,
		},
		Payload: team,
	})
	return nil
}

func NewTeamSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
) TeamSyncer {
	return TeamSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		teamDao:             teamDao,
	}
}
