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

type DeleteSprintParticipant struct {
	logger                 telemetry.Logger
	stateSyncer            *realtime.StateSyncer
	sprintParticipantDao   dao.SprintParticipant
	sprintParticipantDaoV2 daov2.SprintParticipant
	sprintDao              dao.Sprint
	sprintDaoV2            daov2.Sprint
	id                     uint64
	userID                 uint64
	sprintID               uint64
	clientNotifiers        []*realtime.ClientNotifier
	notifiersPrepared      bool
}

var _ realtime.Mutation = (*DeleteSprintParticipant)(nil)

func (d *DeleteSprintParticipant) GetID() uint64 {
	return d.id
}

func (d *DeleteSprintParticipant) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := d.sprintParticipantDaoV2.DeleteSprintParticipant(ct, tx, d.sprintID, d.userID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeleteSprintParticipant) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifiersPrepared {
		return nil
	}
	sprint, err := d.sprintDaoV2.FindSprintByIDWithTx(ct, tx, d.sprintID)
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

func (d *DeleteSprintParticipant) Execute(ct context.Context) *errs.Error {
	err := d.sprintParticipantDao.DeleteSprintParticipant(ct, d.sprintID, d.userID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteSprintParticipant) Undo() *errs.Error {
	return nil
}

func (d *DeleteSprintParticipant) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	sprint, err := d.sprintDao.FindSprintByID(ct, d.sprintID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return d.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (d *DeleteSprintParticipant) GetClientNotifiersV2() []*realtime.ClientNotifier {
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
	sprintParticipantDaoV2 daov2.SprintParticipant,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	userID uint64,
	sprintID uint64,
) *DeleteSprintParticipant {
	return &DeleteSprintParticipant{
		logger:                 logger,
		stateSyncer:            stateSyncer,
		sprintParticipantDao:   sprintParticipantDao,
		sprintParticipantDaoV2: sprintParticipantDaoV2,
		sprintDao:              sprintDao,
		sprintDaoV2:            sprintDaoV2,
		id:                     stateSyncer.NextMutationID(),
		userID:                 userID,
		sprintID:               sprintID,
		notifiersPrepared:      false,
	}
}
