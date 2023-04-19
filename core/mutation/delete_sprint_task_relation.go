package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintTaskRelation struct {
	logger                  telemetry.Logger
	stateSyncer             *realtime.StateSyncer
	sprintTaskRelationDao   dao.SprintTaskRelation
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation
	id                      uint64
	sprintID                uint64
	task                    entity.Task
	clientNotifiers         []*realtime.ClientNotifier
	notifierPrepared        bool
}

var _ realtime.Mutation = (*DeleteSprintTaskRelation)(nil)

func (d *DeleteSprintTaskRelation) GetID() uint64 {
	return d.id
}

func (d *DeleteSprintTaskRelation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := d.sprintTaskRelationDaoV2.DeleteSprintTaskRelation(ct, tx, d.sprintID, d.task.ID)
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

func (d *DeleteSprintTaskRelation) Execute(ct context.Context) *errs.Error {
	err := d.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, d.sprintID, d.task.ID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeleteSprintTaskRelation) Undo() *errs.Error {
	return nil
}

func (d *DeleteSprintTaskRelation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.task.OwningTeamID)
}

func (d *DeleteSprintTaskRelation) GetClientNotifiersV2() []*realtime.ClientNotifier {
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
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation,
	sprintID uint64,
	task entity.Task,
) *DeleteSprintTaskRelation {
	return &DeleteSprintTaskRelation{
		logger:                  logger,
		stateSyncer:             stateSyncer,
		sprintTaskRelationDao:   sprintTaskRelationDao,
		sprintTaskRelationDaoV2: sprintTaskRelationDaoV2,
		id:                      stateSyncer.NextMutationID(),
		sprintID:                sprintID,
		task:                    task,
		notifierPrepared:        false,
	}
}
