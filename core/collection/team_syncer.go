package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const teamCollectionType = "Team"

type TeamSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	teamDao            dao.Team
}

func (t TeamSyncer) CreateAndSyncTeam(team entity.Team) error {
	err := t.teamDao.CreateTeam(team)
	if err != nil {
		return err
	}

	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: teamCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes:     storage.MapAttributes(team),
	})
}

func (t TeamSyncer) UpdateAndSyncTeam(team entity.Team) error {
	err := t.teamDao.UpdateTeam(team)
	if err != nil {
		return err
	}

	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: teamCollectionType,
		MutationType:   storage.UpdateMutationType,
		Attributes:     storage.MapAttributes(team),
	})
}

func NewTeamSyncer(realTimeCollection *storage.RealTimeCollections, teamDao dao.Team) TeamSyncer {
	realTimeCollection.RegisterCollectionType(teamCollectionType)
	return TeamSyncer{
		realTimeCollection: realTimeCollection,
		teamDao:            teamDao,
	}
}
