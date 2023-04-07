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

type CreateSprintTaskRelationMutation struct {
	logger                  telemetry.Logger
	stateSyncer             *realtime.StateSyncer
	sprintTaskRelationDao   dao.SprintTaskRelation
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation
	sprintDao               dao.Sprint
	sprintDaoV2             daov2.Sprint
	id                      uint64
	sprintTaskRelation      entity.SprintTaskRelation
	clientNotifiers         []*realtime.ClientNotifier
	notifiersPrepared       bool
}

var _ realtime.Mutation = (*CreateSprintTaskRelationMutation)(nil)

func (c *CreateSprintTaskRelationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintTaskRelationMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	err := c.sprintTaskRelationDaoV2.CreateSprintTaskRelation(ct, tx, c.sprintTaskRelation)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateSprintTaskRelationMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	sprint, err := c.sprintDao.FindSprintByID(ct, c.sprintTaskRelation.SprintID)
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

func (c *CreateSprintTaskRelationMutation) Execute(ct context.Context) *errs.Error {
	err := c.sprintTaskRelationDao.CreateSprintTaskRelation(ct, c.sprintTaskRelation)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateSprintTaskRelationMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateSprintTaskRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	sprint, err := c.sprintDao.FindSprintByID(ct, c.sprintTaskRelation.SprintID)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (c *CreateSprintTaskRelationMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateSprintTaskRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintTaskRelation,
	}
}

func (c *CreateSprintTaskRelationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateSprintTaskRelationMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	sprintTaskRelation entity.SprintTaskRelation,
) *CreateSprintTaskRelationMutation {
	return &CreateSprintTaskRelationMutation{
		stateSyncer:             stateSyncer,
		sprintTaskRelationDao:   sprintTaskRelationDao,
		sprintTaskRelationDaoV2: sprintTaskRelationDaoV2,
		sprintDao:               sprintDao,
		sprintDaoV2:             sprintDaoV2,
		logger:                  logger,
		id:                      stateSyncer.NextMutationID(),
		sprintTaskRelation:      sprintTaskRelation,
		notifiersPrepared:       false,
	}
}
