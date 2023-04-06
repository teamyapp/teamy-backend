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

type CreateTeamMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamDao          dao.Team
	teamDaoV2        daov2.Team
	id               uint64
	team             entity.Team
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTeamMutation)(nil)

func (c *CreateTeamMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.teamDaoV2.CreateTeam(ct, tx, c.team)
}

func (c *CreateTeamMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	var err *errs.Error
	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.team.ID)
	if err != nil {
		return err
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateTeamMutation) Execute(ct context.Context) *errs.Error {
	err := c.teamDao.CreateTeam(ct, c.team)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateTeamMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.team.ID)
}

func (c *CreateTeamMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.team,
	}
}

func (c *CreateTeamMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamMutation(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	teamDaoV2 daov2.Team,
	team entity.Team,
) *CreateTeamMutation {
	return &CreateTeamMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamDao:          teamDao,
		teamDaoV2:        teamDaoV2,
		id:               stateSyncer.NextMutationID(),
		team:             team,
		notifierPrepared: false,
	}
}
