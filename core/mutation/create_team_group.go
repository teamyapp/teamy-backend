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

type CreateTeamGroup struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	teamGroupDao      dao.TeamGroup
	id                uint64
	teamGroup         entity.TeamGroup
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateTeamGroup)(nil)

func (c *CreateTeamGroup) GetID() uint64 {
	return c.id
}

func (c *CreateTeamGroup) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.teamGroupDao.CreateGroup(ct, tx, c.teamGroup)
}

func (c *CreateTeamGroup) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamGroup.TeamID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateTeamGroup) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamGroup) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeamGroup) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamGroupCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.teamGroup,
	}
}

func (c *CreateTeamGroup) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamGroup(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamGroupDao dao.TeamGroup,
	teamGroup entity.TeamGroup,
) *CreateTeamGroup {
	return &CreateTeamGroup{
		logger:            logger,
		stateSyncer:       stateSyncer,
		teamGroupDao:      teamGroupDao,
		id:                stateSyncer.NextMutationID(),
		teamGroup:         teamGroup,
		notifiersPrepared: false,
	}
}
