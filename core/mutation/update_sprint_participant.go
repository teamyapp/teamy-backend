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

type UpdateSprintParticipant struct {
	logger               telemetry.Logger
	stateSyncer          *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
	id                   uint64
	sprintParticipant    entity.SprintParticipant
	clientNotifiers      []*realtime.ClientNotifier
	notifierPrepared     bool
}

var _ realtime.Mutation = (*UpdateSprintParticipant)(nil)

func (u *UpdateSprintParticipant) GetID() uint64 {
	return u.id
}

func (u *UpdateSprintParticipant) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := u.sprintParticipantDao.UpdateSprintParticipant(ct, tx, u.sprintParticipant)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (u *UpdateSprintParticipant) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	sprint, internalErr := u.sprintDao.FindSprintByIDWithTx(ct, tx, u.sprintParticipant.SprintID)
	if internalErr != nil {
		return internalErr
	}

	u.clientNotifiers, internalErr = u.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateSprintParticipant) Undo() *errs.Error {
	return nil
}

func (u *UpdateSprintParticipant) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateSprintParticipant) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.sprintParticipant,
	}
}

func (u *UpdateSprintParticipant) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateSprintParticipant(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	sprintParticipant entity.SprintParticipant,
) *UpdateSprintParticipant {
	return &UpdateSprintParticipant{
		logger:               logger,
		stateSyncer:          stateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
		id:                   stateSyncer.NextMutationID(),
		sprintParticipant:    sprintParticipant,
		notifierPrepared:     false,
	}
}
