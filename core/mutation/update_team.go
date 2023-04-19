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

type UpdateTeam struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamDao          dao.Team
	teamDaoV2        daov2.Team
	id               uint64
	team             entity.Team
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateTeam)(nil)

func (u *UpdateTeam) GetID() uint64 {
	return u.id
}

func (u *UpdateTeam) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.teamDaoV2.UpdateTeam(ct, tx, u.team)
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

func (u *UpdateTeam) Execute(ct context.Context) *errs.Error {
	err := u.teamDao.UpdateTeam(ct, u.team)
	if err != nil {
		u.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateTeam) Undo() *errs.Error {
	return nil
}

func (u *UpdateTeam) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.team.ID)
}

func (u *UpdateTeam) GetClientNotifiersV2() []*realtime.ClientNotifier {
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
	teamDaoV2 daov2.Team,
	team entity.Team,
) *UpdateTeam {
	return &UpdateTeam{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamDao:          teamDao,
		teamDaoV2:        teamDaoV2,
		id:               stateSyncer.NextMutationID(),
		team:             team,
		notifierPrepared: false,
	}
}
