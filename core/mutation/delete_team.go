package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeam struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamDao          dao.Team
	id               uint64
	teamID           uint64
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTeam)(nil)

func (d *DeleteTeam) GetID() uint64 {
	return d.id
}

func (d *DeleteTeam) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.teamDao.DeleteTeam(ct, tx, d.teamID)
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

func (d *DeleteTeam) GetClientNotifiers() []*realtime.ClientNotifier {
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
	teamDao dao.Team,
	teamID uint64,
) *DeleteTeam {
	return &DeleteTeam{
		logger:      logger,
		stateSyncer: stateSyncer,
		teamDao:     teamDao,
		id:          stateSyncer.NextMutationID(),
		teamID:      teamID,
	}
}
