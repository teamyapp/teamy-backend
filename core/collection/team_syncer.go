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

func (t TeamSyncer) CreateAndSyncTeam(ct context.Context, tx realtime.Transaction, team entity.Team) error {
	err := t.teamDao.CreateTeam(ct, team)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        team,
	})
	return nil
}

func (t TeamSyncer) UpdateAndSyncTeam(ct context.Context, tx realtime.Transaction, team entity.Team) error {
	err := t.teamDao.UpdateTeam(ct, team)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        team,
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
