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

type UpdateTeam struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamDao          dao.Team
	id               uint64
	team             entity.Team
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateTeam)(nil)

func (u *UpdateTeam) GetID() uint64 {
	return u.id
}

func (u *UpdateTeam) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.teamDao.UpdateTeam(ct, tx, u.team)
}

func (u *UpdateTeam) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	var err *errs.Error
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.team.ID)
	if err != nil {
		return err
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateTeam) Undo() *errs.Error {
	return nil
}

func (u *UpdateTeam) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateTeam) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.team,
	}
}

func (u *UpdateTeam) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTeam(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	team entity.Team,
) *UpdateTeam {
	return &UpdateTeam{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamDao:          teamDao,
		id:               stateSyncer.NextMutationID(),
		team:             team,
		notifierPrepared: false,
	}
}
