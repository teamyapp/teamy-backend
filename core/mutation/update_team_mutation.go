package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamDao       dao.Team
	id            uint64
	team          entity.Team
}

var _ realtime.Mutation = (*UpdateTeamMutation)(nil)

func (u *UpdateTeamMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateTeamMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTeamMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTeamMutation) Execute(ct context.Context) *errs.Error {
	err := u.teamDao.UpdateTeam(ct, u.team)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateTeamMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.team.ID)
}

func (u *UpdateTeamMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.team,
	}
}

func (u *UpdateTeamMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTeamMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	team entity.Team,
) *UpdateTeamMutation {
	return &UpdateTeamMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		id:            stateSyncer.NextMutationID(),
		team:          team,
	}
}
