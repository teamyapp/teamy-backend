package collection

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TeamMemberSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	teamMemberDao       dao.TeamMember
}

func (t TeamMemberSyncer) CreateAndSyncTeamMember(teamID uint64, userID uint64) error {
	err := t.teamMemberDao.CreateTeamMember(teamID, userID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
	},
	)
	return nil
}

func (t TeamMemberSyncer) DeleteAndSyncTeamMember(teamID uint64, userID uint64) error {
	err := t.teamMemberDao.DeleteTeamMember(teamID, userID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
	},
	)
	return nil
}

func NewTeamMemberSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
) TeamMemberSyncer {
	return TeamMemberSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		teamMemberDao:       teamMemberDao,
	}
}
