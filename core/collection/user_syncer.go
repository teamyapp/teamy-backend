package collection

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UserSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	userDao             dao.User
	teamMemberDao       dao.TeamMember
}

func (u UserSyncer) UpdateAndSyncUser(user entity.User) error {
	err := u.userDao.UpdateUser(user)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	teamIDs, err := u.teamMemberDao.FindTeamIDsByUserID(user.ID)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	u.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.UserCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs:        teamIDs,
		Payload:        user,
	},
	)
	return nil
}

func NewUserSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	userDao dao.User,
	teamMemberDao dao.TeamMember,
) UserSyncer {
	return UserSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		userDao:             userDao,
		teamMemberDao:       teamMemberDao,
	}
}
