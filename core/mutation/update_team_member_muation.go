package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMemberMutation struct {
	dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	id            uint64
	teamMember    entity.TeamMember
}

func (u *UpdateTeamMemberMutation) GetID() uint64 {
	return u.id
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

func (u *UpdateTeamMemberMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewUpdateTeamMemberMutation(
	dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	teamMember entity.TeamMember,
) *UpdateTeamMemberMutation {
	return &UpdateTeamMemberMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		id:            stateSyncer.NextMutationID(),
		teamMember:    teamMember,
	}
}
