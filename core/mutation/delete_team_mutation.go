package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeamMutation struct {
	dataCollector    telemetry.DataCollector
	stateSyncer      *realtime.StateSyncer
	teamDao          dao.Team
	teamDaoV2        daov2.Team
	id               uint64
	teamID           uint64
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTeamMutation)(nil)

func (d *DeleteTeamMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTeamMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.teamDaoV2.DeleteTeam(ct, tx, d.teamID)
}

func (d *DeleteTeamMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	var err *errs.Error
	d.clientNotifiers, err = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
	if err != nil {
		return err
	}

	d.notifierPrepared = true
	return nil
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
	return d.clientNotifiers
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
	teamDaoV2 daov2.Team,
	teamID uint64,
) *DeleteTeamMutation {
	return &DeleteTeamMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		teamDaoV2:     teamDaoV2,
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
	}
}
