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
	tx realtime.Transaction,
	teamMember entity.TeamMember,
) error {
	err := t.teamMemberDao.CreateTeamMember(ct, teamMember)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        teamMember,
	})
	return nil
}

func (t TeamMemberSyncer) UpdateAndSyncTeamMember(
	ct context.Context,
	tx realtime.Transaction,
	teamMember entity.TeamMember,
) error {
	err := t.teamMemberDao.UpdateTeamMember(ct, teamMember)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        teamMember,
	})
	return nil
}

func (t TeamMemberSyncer) DeleteAndSyncTeamMember(
	ct context.Context,
	tx realtime.Transaction,
	teamID uint64,
	userID uint64) error {
	err := t.teamMemberDao.DeleteTeamMember(ct, teamID, userID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.DeleteMutationType,
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
