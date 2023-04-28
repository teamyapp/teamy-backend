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

type CreateTeam struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamDaoV2        daov2.Team
	id               uint64
	team             entity.Team
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateTeam)(nil)

func (c *CreateTeam) GetID() uint64 {
	return c.id
}

func (c *CreateTeam) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.teamDaoV2.CreateTeam(ct, tx, c.team)
}

func (c *CreateTeam) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (c *CreateTeam) Undo() *errs.Error {
	return nil
}

func (c *CreateTeam) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateTeam) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.team,
	}
}

func (c *CreateTeam) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeam(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamDaoV2 daov2.Team,
	team entity.Team,
) *CreateTeam {
	return &CreateTeam{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamDaoV2:        teamDaoV2,
		id:               stateSyncer.NextMutationID(),
		team:             team,
		notifierPrepared: false,
	}
}
