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

type UpdateUserMutation struct {
	dataCollector     telemetry.DataCollector
	stateSyncer       *realtime.StateSyncer
	userDao           dao.User
	userDaoV2         daov2.User
	teamMemberDaoV2   daov2.TeamMember
	id                uint64
	user              entity.User
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*UpdateUserMutation)(nil)

func (u *UpdateUserMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateUserMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.userDaoV2.UpdateUser(ct, tx, u.user)
}

func (u *UpdateUserMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifiersPrepared {
		return nil
	}

	teamIDs, err := u.teamMemberDaoV2.FindTeamIDsByUserIDWithTx(ct, tx, u.user.ID)
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamIDs(ct, teamIDs)
	u.notifiersPrepared = true
	return err
}

func (u *UpdateUserMutation) Execute(ct context.Context) *errs.Error {
	err := u.userDao.UpdateUser(ct, u.user)
	if err != nil {
		return err
	}

	return nil
}

func (u *UpdateUserMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateUserMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return u.stateSyncer.GetAllClientNotifiersByUserID(ct, u.user.ID)
}

func (u *UpdateUserMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateUserMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.UserCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.user,
	}
}

func (u *UpdateUserMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateUserMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	userDao dao.User,
	userDaoV2 daov2.User,
	teamMemberDaoV2 daov2.TeamMember,
	user entity.User,
) *UpdateUserMutation {
	return &UpdateUserMutation{
		dataCollector:     dataCollector,
		stateSyncer:       stateSyncer,
		userDao:           userDao,
		userDaoV2:         userDaoV2,
		teamMemberDaoV2:   teamMemberDaoV2,
		id:                stateSyncer.NextMutationID(),
		user:              user,
		notifiersPrepared: false,
	}
}
