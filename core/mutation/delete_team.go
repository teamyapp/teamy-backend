package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeam struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamDaoV2        daov2.Team
	id               uint64
	teamID           uint64
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTeam)(nil)

func (d *DeleteTeam) GetID() uint64 {
	return d.id
}

func (d *DeleteTeam) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.teamDaoV2.DeleteTeam(ct, tx, d.teamID)
}

func (d *DeleteTeam) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (d *DeleteTeam) Undo() *errs.Error {
	return nil
}

func (d *DeleteTeam) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTeam) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.teamID,
	}
}

func (d *DeleteTeam) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTeam(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamDaoV2 daov2.Team,
	teamID uint64,
) *DeleteTeam {
	return &DeleteTeam{
		logger:      logger,
		stateSyncer: stateSyncer,
		teamDaoV2:   teamDaoV2,
		id:          stateSyncer.NextMutationID(),
		teamID:      teamID,
	}
}
