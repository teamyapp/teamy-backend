package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMemberMutation struct {
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	dataCollector obs.DataCollector
	id            uint64
	teamMember    entity.TeamMember
}

func (c *UpdateTeamMemberMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateTeamMemberMutation) Execute(ct context.Context) error {
	err := u.teamMemberDao.UpdateTeamMember(ct, u.teamMember)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateTeamMemberMutation) Undo() error {
	return nil
}

func (u *UpdateTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamMember.TeamID)
}

func (u *UpdateTeamMemberMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.teamMember,
	}
}

func NewUpdateTeamMemberMutation(
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	dataCollector obs.DataCollector,
	teamMember entity.TeamMember) *UpdateTeamMemberMutation {
	return &UpdateTeamMemberMutation{
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		teamMember:    teamMember,
	}
}
