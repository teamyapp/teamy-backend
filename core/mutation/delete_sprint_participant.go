package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintParticipant struct {
	logger               telemetry.Logger
	stateSyncer          *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
	id                   uint64
	userID               uint64
	sprintID             uint64
	clientNotifiers      []*realtime.ClientNotifier
	notifiersPrepared    bool
}

var _ realtime.Mutation = (*DeleteSprintParticipant)(nil)

func (d *DeleteSprintParticipant) GetID() uint64 {
	return d.id
}

func (d *DeleteSprintParticipant) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := d.sprintParticipantDao.DeleteSprintParticipant(ct, tx, d.sprintID, d.userID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeleteSprintParticipant) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifiersPrepared {
		return nil
	}
	sprint, err := d.sprintDao.FindSprintByIDWithTx(ct, tx, d.sprintID)
	if err != nil {
		return err
	}

	d.clientNotifiers, err = d.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
	if err != nil {
		return err
	}

	d.notifiersPrepared = true
	return nil
}

func (d *DeleteSprintParticipant) Undo() *errs.Error {
	return nil
}

func (d *DeleteSprintParticipant) GetClientNotifiers() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteSprintParticipant) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			SprintID uint64
			UserID   uint64
		}{
			SprintID: d.sprintID,
			UserID:   d.userID,
		},
	}
}

func (d *DeleteSprintParticipant) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteSprintParticipant(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	userID uint64,
	sprintID uint64,
) *DeleteSprintParticipant {
	return &DeleteSprintParticipant{
		logger:               logger,
		stateSyncer:          stateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
		id:                   stateSyncer.NextMutationID(),
		userID:               userID,
		sprintID:             sprintID,
		notifiersPrepared:    false,
	}
}
