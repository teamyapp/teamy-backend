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

type UpdateUser struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	userDao           dao.User
	teamMemberDao     dao.TeamMember
	id                uint64
	user              entity.User
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*UpdateUser)(nil)

func (u *UpdateUser) GetID() uint64 {
	return u.id
}

func (u *UpdateUser) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.userDao.UpdateUser(ct, tx, u.user)
}

func (u *UpdateUser) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifiersPrepared {
		return nil
	}

	teamIDs, err := u.teamMemberDao.FindTeamIDsByUserIDWithTx(ct, tx, u.user.ID)
	if err != nil {
		return err
	}

	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamIDs(ct, teamIDs)
	if err != nil {
		return err
	}

	u.notifiersPrepared = true
	return err
}

func (u *UpdateUser) Undo() *errs.Error {
	return nil
}

func (u *UpdateUser) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateUser) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.UserCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.user,
	}
}

func (u *UpdateUser) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateUser(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	userDao dao.User,
	teamMemberDao dao.TeamMember,
	user entity.User,
) *UpdateUser {
	return &UpdateUser{
		logger:            logger,
		stateSyncer:       stateSyncer,
		userDao:           userDao,
		teamMemberDao:     teamMemberDao,
		id:                stateSyncer.NextMutationID(),
		user:              user,
		notifiersPrepared: false,
	}
}
