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

type CreateSprintParticipantMutation struct {
	logger                 telemetry.Logger
	stateSyncer            *realtime.StateSyncer
	sprintParticipantDao   dao.SprintParticipant
	sprintParticipantDaoV2 daov2.SprintParticipant
	sprintDao              dao.Sprint
	sprintDaoV2            daov2.Sprint
	id                     uint64
	sprintParticipant      entity.SprintParticipant
	clientNotifiers        []*realtime.ClientNotifier
	notifiersPrepared      bool
}

var _ realtime.Mutation = (*CreateSprintParticipantMutation)(nil)

func (c *CreateSprintParticipantMutation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintParticipantMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.sprintParticipantDaoV2.CreateSprintParticipant(ct, tx, c.sprintParticipant)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateSprintParticipantMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	sprint, err := c.sprintDaoV2.FindSprintByIDWithTx(ct, tx, c.sprintParticipant.SprintID)
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

func (c *CreateSprintParticipantMutation) Execute(ct context.Context) *errs.Error {
	err := c.sprintParticipantDao.CreateSprintParticipant(ct, c.sprintParticipant)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateSprintParticipantMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateSprintParticipantMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	sprint, err := c.sprintDao.FindSprintByID(ct, c.sprintParticipant.SprintID)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (c *CreateSprintParticipantMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateSprintParticipantMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintParticipant,
	}
}

func (c *CreateSprintParticipantMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateSprintParticipantMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	sprintParticipant entity.SprintParticipant,
) *CreateSprintParticipantMutation {
	return &CreateSprintParticipantMutation{
		logger:                 logger,
		stateSyncer:            stateSyncer,
		sprintParticipantDao:   sprintParticipantDao,
		sprintParticipantDaoV2: sprintParticipantDaoV2,
		sprintDao:              sprintDao,
		sprintDaoV2:            sprintDaoV2,
		id:                     stateSyncer.NextMutationID(),
		sprintParticipant:      sprintParticipant,
		notifiersPrepared:      false,
	}
}
