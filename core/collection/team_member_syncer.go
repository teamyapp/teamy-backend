package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TeamMemberSyncer struct {
	realTimeStateSyncer *realtime.StateSyncer
	teamMemberDao       dao.TeamMember
}

func (t TeamMemberSyncer) CreateAndSyncTeamMember(teamID uint64, userID uint64) error {
	err := t.teamMemberDao.CreateTeamMember(teamID, userID)
	if err != nil {
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			teamID,
		},
		Payload: entity.TeamMember{
			TeamID: teamID,
			UserID: userID,
		},
	})
	return nil
}

func (t TeamMemberSyncer) DeleteAndSyncTeamMember(teamID uint64, userID uint64) error {
	err := t.teamMemberDao.DeleteTeamMember(teamID, userID)
	if err != nil {
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			teamID,
		},
		Payload: entity.TeamMember{
			TeamID: teamID,
			UserID: userID,
		},
	})
	return nil
}

func NewTeamMemberSyncer(realTimeStateSyncer *realtime.StateSyncer, teamMemberDao dao.TeamMember) TeamMemberSyncer {
	return TeamMemberSyncer{
		realTimeStateSyncer: realTimeStateSyncer,
		teamMemberDao:       teamMemberDao,
	}
}
