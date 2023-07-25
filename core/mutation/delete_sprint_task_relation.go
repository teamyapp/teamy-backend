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

type DeleteSprintTaskRelation struct {
	logger                telemetry.Logger
	stateSyncer           *realtime.StateSyncer
	sprintTaskRelationDao dao.SprintTaskRelation
	id                    uint64
	sprintID              uint64
	task                  entity.Task
	clientNotifiers       []*realtime.ClientNotifier
	notifierPrepared      bool
}

var _ realtime.Mutation = (*DeleteSprintTaskRelation)(nil)

func (d *DeleteSprintTaskRelation) GetID() uint64 {
	return d.id
}

func (d *DeleteSprintTaskRelation) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := d.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, tx, d.sprintID, d.task.ID)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (d *DeleteSprintTaskRelation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	d.clientNotifiers, internalErr = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteSprintTaskRelation) Undo() *errs.Error {
	return nil
}

func (d *DeleteSprintTaskRelation) GetClientNotifiers() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteSprintTaskRelation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			SprintID uint64
			TaskID   uint64
		}{
			SprintID: d.sprintID,
			TaskID:   d.task.ID,
		},
	}
}

func (d *DeleteSprintTaskRelation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteSprintTaskRelation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintID uint64,
	task entity.Task,
) *DeleteSprintTaskRelation {
	return &DeleteSprintTaskRelation{
		logger:                logger,
		stateSyncer:           stateSyncer,
		sprintTaskRelationDao: sprintTaskRelationDao,
		id:                    stateSyncer.NextMutationID(),
		sprintID:              sprintID,
		task:                  task,
		notifierPrepared:      false,
	}
}
