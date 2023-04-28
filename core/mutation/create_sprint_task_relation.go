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

type CreateSprintTaskRelation struct {
	logger                  telemetry.Logger
	stateSyncer             *realtime.StateSyncer
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation
	sprintDaoV2             daov2.Sprint
	id                      uint64
	sprintTaskRelation      entity.SprintTaskRelation
	clientNotifiers         []*realtime.ClientNotifier
	notifiersPrepared       bool
}

var _ realtime.Mutation = (*CreateSprintTaskRelation)(nil)

func (c *CreateSprintTaskRelation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintTaskRelation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.sprintTaskRelationDaoV2.CreateSprintTaskRelation(ct, tx, c.sprintTaskRelation)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateSprintTaskRelation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	sprint, err := c.sprintDaoV2.FindSprintByIDWithTx(ct, tx, c.sprintTaskRelation.SprintID)
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

func (c *CreateSprintTaskRelation) Undo() *errs.Error {
	return nil
}

func (c *CreateSprintTaskRelation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateSprintTaskRelation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintTaskRelation,
	}
}

func (c *CreateSprintTaskRelation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateSprintTaskRelation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation,
	sprintDaoV2 daov2.Sprint,
	sprintTaskRelation entity.SprintTaskRelation,
) *CreateSprintTaskRelation {
	return &CreateSprintTaskRelation{
		stateSyncer:             stateSyncer,
		sprintTaskRelationDaoV2: sprintTaskRelationDaoV2,
		sprintDaoV2:             sprintDaoV2,
		logger:                  logger,
		id:                      stateSyncer.NextMutationID(),
		sprintTaskRelation:      sprintTaskRelation,
		notifiersPrepared:       false,
	}
}
