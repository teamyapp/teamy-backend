package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateUserMutation struct {
	stateSyncer   *realtime.StateSyncer
	teamMemberDao dao.TeamMember
	userDao       dao.User
	dataCollector obs.DataCollector
	id            uint64
	user          entity.User
}

func (c *UpdateUserMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateUserMutation) Execute(ct context.Context) error {
	err := u.userDao.UpdateUser(ct, u.user)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateUserMutation) Undo() error {
	return nil
}

func (u *UpdateUserMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByUserID(ct, u.user.ID)
}

func (u *UpdateUserMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.UserCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.user,
	}
}

func NewUpdateUserMutation(
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	userDao dao.User,
	dataCollector obs.DataCollector,
	user entity.User) *UpdateUserMutation {
	return &UpdateUserMutation{
		stateSyncer:   stateSyncer,
		teamMemberDao: teamMemberDao,
		userDao:       userDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		user:          user,
	}
}
