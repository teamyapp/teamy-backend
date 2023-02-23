package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMemberMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	id            uint64
	teamMember    entity.TeamMember
}

var _ realtime.Mutation = (*UpdateTeamMemberMutation)(nil)

func (u *UpdateTeamMemberMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateTeamMemberMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTeamMemberMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTeamMemberMutation) Execute(ct context.Context) *errs.Error {
	err := u.teamMemberDao.UpdateTeamMember(ct, u.teamMember)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateTeamMemberMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
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

func (u *UpdateTeamMemberMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTeamMemberMutation(
	dataCollector telemetry.DataCollector,
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
