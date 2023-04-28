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

type DeleteTaskAwaitForRelation struct {
	logger                    telemetry.Logger
	stateSyncer               *realtime.StateSyncer
	taskAwaitForRelationDaoV2 daov2.TaskAwaitForRelation
	id                        uint64
	awaitingTask              entity.Task
	awaitForTaskID            uint64
	clientNotifiers           []*realtime.ClientNotifier
	notifierPrepared          bool
}

var _ realtime.Mutation = (*DeleteTaskAwaitForRelation)(nil)

func (d *DeleteTaskAwaitForRelation) GetID() uint64 {
	return d.id
}

func (d *DeleteTaskAwaitForRelation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := d.taskAwaitForRelationDaoV2.DeleteRelation(ct, tx, d.awaitingTask.ID, d.awaitForTaskID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeleteTaskAwaitForRelation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	var err *errs.Error
	d.clientNotifiers, err = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.awaitingTask.OwningTeamID)
	if err != nil {
		return err
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteTaskAwaitForRelation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTaskAwaitForRelation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTaskAwaitForRelation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			AwaitingTaskID uint64
			AwaitForTaskID uint64
		}{
			AwaitingTaskID: d.awaitingTask.ID,
			AwaitForTaskID: d.awaitForTaskID,
		},
	}
}

func (d *DeleteTaskAwaitForRelation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTaskAwaitForRelation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDaoV2 daov2.TaskAwaitForRelation,
	awaitingTask entity.Task,
	awaitForTaskID uint64,
) *DeleteTaskAwaitForRelation {
	return &DeleteTaskAwaitForRelation{
		logger:                    logger,
		stateSyncer:               stateSyncer,
		taskAwaitForRelationDaoV2: taskAwaitForRelationDaoV2,
		id:                        stateSyncer.NextMutationID(),
		awaitingTask:              awaitingTask,
		awaitForTaskID:            awaitForTaskID,
		notifierPrepared:          false,
	}
}
