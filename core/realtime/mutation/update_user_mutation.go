package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateUserMutation struct {
	id            uint64
	stateSyncer   *realtime.StateSyncer
	user          entity.User
	teamMemberDao dao.TeamMember
	userDao       dao.User
	dataCollector obs.DataCollector
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
	teamIDs, err := u.teamMemberDao.FindTeamIDsByUserID(ct, u.user.ID)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	clientNotifiers := make([]*realtime.ClientNotifier, 0)

	for _, teamID := range teamIDs {
		teamClientNotifiers, err := u.stateSyncer.GetClientNotifiersByTeamID(ct, teamID)
		if err != nil {
			u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return []*realtime.ClientNotifier{}, err
		}
		clientNotifiers = append(clientNotifiers, teamClientNotifiers...)
	}

	return clientNotifiers, nil
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
	user entity.User,
	teamMemberDao dao.TeamMember,
	userDao dao.User,
	dataCollector obs.DataCollector) *UpdateUserMutation {
	return &UpdateUserMutation{
		id:            stateSyncer.NextMutationID(),
		stateSyncer:   stateSyncer,
		user:          user,
		teamMemberDao: teamMemberDao,
		userDao:       userDao,
		dataCollector: dataCollector,
	}
}
