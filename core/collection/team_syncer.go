package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TeamSyncer struct {
	realTimeStateSyncer *realtime.StateSyncer
	teamDao             dao.Team
}

func (t TeamSyncer) CreateAndSyncTeam(team entity.Team) error {
	err := t.teamDao.CreateTeam(team)
	if err != nil {
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

func (t TeamSyncer) UpdateAndSyncTeam(team entity.Team) error {
	err := t.teamDao.UpdateTeam(team)
	if err != nil {
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

func NewTeamSyncer(realTimeStateSyncer *realtime.StateSyncer, teamDao dao.Team) TeamSyncer {
	return TeamSyncer{
		realTimeStateSyncer: realTimeStateSyncer,
		teamDao:             teamDao,
	}
}
