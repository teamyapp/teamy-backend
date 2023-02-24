package mutation

import (
	"context"
	"database/sql"

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

var _ realtime.Mutation = (*DeleteTeamMutation)(nil)

func (d *DeleteTeamMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTeamMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTeamMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTeamMutation) Execute(ct context.Context) *errs.Error {
	err := d.teamDao.DeleteTeam(ct, d.teamID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteTeamMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteTeamMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.teamID,
	}
}

func (d *DeleteTeamMutation) CleanUp(ct context.Context) *errs.Error {
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
