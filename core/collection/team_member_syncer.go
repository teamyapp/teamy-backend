package collection

import (
	"context"

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

func (t TeamMemberSyncer) CreateAndSyncTeamMember(
	ct context.Context,
	teamMember entity.TeamMember,
) error {
	err := t.teamMemberDao.CreateTeamMember(ct, teamMember)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			teamMember.TeamID,
		},
		Payload: teamMember,
	})
	return nil
}

func (t TeamMemberSyncer) DeleteAndSyncTeamMember(ct context.Context, teamID uint64, userID uint64) error {
	err := t.teamMemberDao.DeleteTeamMember(ct, teamID, userID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
