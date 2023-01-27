package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateUserMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	userDao       dao.User
	id            uint64
	user          entity.User
}

func (u *UpdateUserMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateUserMutation) Execute(ct context.Context) error {
	err := u.userDao.UpdateUser(ct, u.user)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateUserMutation) Undo() error {
	return nil
}

func (u *UpdateUserMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetAllClientNotifiersByUserID(ct, u.user.ID)
}

func (u *UpdateUserMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.UserCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.user,
	}
}

func (u *UpdateUserMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewUpdateUserMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	userDao dao.User,
	user entity.User,
) *UpdateUserMutation {
	return &UpdateUserMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		userDao:       userDao,
		id:            stateSyncer.NextMutationID(),
		user:          user,
	}
}
