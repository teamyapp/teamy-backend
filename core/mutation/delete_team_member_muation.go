package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeamMemberMutation struct {
	dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	id            uint64
	teamID        uint64
	userID        uint64
}

func (c *DeleteTeamMemberMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteTeamMemberMutation) Execute(ct context.Context) error {
	err := d.teamMemberDao.DeleteTeamMember(ct, d.teamID, d.userID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteTeamMemberMutation) Undo() error {
	return nil
}

func (d *DeleteTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteTeamMemberMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: entity.TeamMember{
			TeamID: d.teamID,
			UserID: d.userID,
		},
	}
}

func (d *DeleteTeamMemberMutation) CleanUp(ct context.Context) error {
	teamNotifier, err := d.stateSyncer.GetTeamNotifier(ct, d.teamID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	teamNotifier.UnregisterUserNotifier(d.userID)
	return nil
}

func NewDeleteTeamMemberMutation(
	dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	teamID uint64,
	userID uint64,
) *DeleteTeamMemberMutation {
	return &DeleteTeamMemberMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		userID:        userID,
	}
}
