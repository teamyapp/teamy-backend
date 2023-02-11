package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeamMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamDao       dao.Team
	id            uint64
	teamID        uint64
}

func (u *DeleteTeamMutation) GetID() uint64 {
	return u.id
}

func (u *DeleteTeamMutation) Execute(ct context.Context) *errs.Error {
	err := u.teamDao.DeleteTeam(ct, u.teamID)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *DeleteTeamMutation) Undo() *errs.Error {
	return nil
}

func (u *DeleteTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamID)
}

func (u *DeleteTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        u.teamID,
	}
}

func (u *DeleteTeamMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTeamMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	teamID uint64,
) *DeleteTeamMutation {
	return &DeleteTeamMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
	}
}
