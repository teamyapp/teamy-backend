package collection

import (
	"strconv"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const teamMemberCollectionType = "TeamMember"

type TeamMemberSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	teamMemberDao      dao.TeamMember
}

func (t TeamMemberSyncer) CreateAndSyncTeamMember(teamID uint64, userID uint64) error {
	err := t.teamMemberDao.CreateTeamMember(teamID, userID)
	if err != nil {
		return err
	}

	teamIDStr := strconv.FormatUint(teamID, 10)
	userIDStr := strconv.FormatUint(userID, 10)
	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: teamMemberCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes: map[string]*string{
			"TeamID": &teamIDStr,
			"UserID": &userIDStr,
		},
	})
}

func (t TeamMemberSyncer) DeleteAndSyncTeamMember(teamID uint64, userID uint64) error {
	err := t.teamMemberDao.DeleteTeamMember(teamID, userID)
	if err != nil {
		return err
	}

	teamIDStr := strconv.FormatUint(teamID, 10)
	userIDStr := strconv.FormatUint(userID, 10)
	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: teamMemberCollectionType,
		MutationType:   storage.DeleteMutationType,
		Attributes: map[string]*string{
			"TeamID": &teamIDStr,
			"UserID": &userIDStr,
		},
	})
}

func NewTeamMemberSyncer(realTimeCollection *storage.RealTimeCollections, teamMemberDao dao.TeamMember) TeamMemberSyncer {
	realTimeCollection.RegisterCollectionType(teamMemberCollectionType)
	return TeamMemberSyncer{
		realTimeCollection: realTimeCollection,
		teamMemberDao:      teamMemberDao,
	}
}
