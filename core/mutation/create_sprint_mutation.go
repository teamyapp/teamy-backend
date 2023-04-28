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

type CreateSprintMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	sprintDaoV2      daov2.Sprint
	id               uint64
	sprint           entity.Sprint
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateSprintMutation)(nil)

func (c *CreateSprintMutation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := c.sprintDaoV2.CreateSprint(ct, tx, c.sprint)
	return internalErr
}

func (c *CreateSprintMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.sprint.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateSprintMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateSprintMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateSprintMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprint,
	}
}

func (c *CreateSprintMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateSprintMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintDaoV2 daov2.Sprint,
	sprint entity.Sprint,
) *CreateSprintMutation {
	return &CreateSprintMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		sprintDaoV2:      sprintDaoV2,
		id:               stateSyncer.NextMutationID(),
		sprint:           sprint,
		notifierPrepared: false,
	}
}
