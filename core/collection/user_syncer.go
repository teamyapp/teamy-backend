package collection

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UserSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	userDao             dao.User
	teamMemberDao       dao.TeamMember
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
