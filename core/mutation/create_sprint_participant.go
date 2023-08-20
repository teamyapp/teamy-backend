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

type CreateSprintParticipant struct {
	logger               telemetry.Logger
	stateSyncer          *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
	id                   uint64
	sprintParticipant    entity.SprintParticipant
	clientNotifiers      []*realtime.ClientNotifier
	notifiersPrepared    bool
}

var _ realtime.Mutation = (*CreateSprintParticipant)(nil)

func (c *CreateSprintParticipant) GetID() uint64 {
	return c.id
}

func (c *CreateSprintParticipant) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.sprintParticipantDao.CreateSprintParticipant(ct, tx, c.sprintParticipant)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateSprintParticipant) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	sprint, err := c.sprintDao.FindSprintByIDWithTx(ct, tx, c.sprintParticipant.SprintID)
	if err != nil {
		return err
	}

	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateSprintParticipant) Undo() *errs.Error {
	return nil
}

func (c *CreateSprintParticipant) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateSprintParticipant) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintParticipant,
	}
}

func (c *CreateSprintParticipant) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateSprintParticipant(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	sprintParticipant entity.SprintParticipant,
) *CreateSprintParticipant {
	return &CreateSprintParticipant{
		logger:               logger,
		stateSyncer:          stateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
		id:                   stateSyncer.NextMutationID(),
		sprintParticipant:    sprintParticipant,
		notifiersPrepared:    false,
	}
}
