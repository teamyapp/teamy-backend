package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprint struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	sprintDaoV2      daov2.Sprint
	id               uint64
	sprint           entity.Sprint
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTask)(nil)

func (d *DeleteSprint) GetID() uint64 {
	return d.id
}

func (d *DeleteSprint) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := d.sprintDaoV2.DeleteSprint(ct, tx, d.sprint.ID)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (d *DeleteSprint) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	d.clientNotifiers, internalErr = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.sprint.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteSprint) Execute(ct context.Context) *errs.Error {
	return nil
}

func (d *DeleteSprint) Undo() *errs.Error {
	return nil
}

func (d *DeleteSprint) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.sprint.OwningTeamID)
}

func (d *DeleteSprint) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteSprint) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.SprintCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.sprint.ID,
	}
}

func (d *DeleteSprint) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteSprint(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintDaoV2 daov2.Sprint,
	sprint entity.Sprint,
) *DeleteSprint {
	return &DeleteSprint{
		logger:           logger,
		stateSyncer:      stateSyncer,
		sprintDaoV2:      sprintDaoV2,
		id:               stateSyncer.NextMutationID(),
		sprint:           sprint,
		notifierPrepared: false,
	}
}
