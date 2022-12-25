package gql

import (
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Dependencies struct {
	dataCollector           obs.DataCollector
	cloudClientRegistry     *api.ClientRegistry
	userDao                 dao.User
	teamDao                 dao.Team
	taskDao                 dao.Task
	teamMemberDao           dao.TeamMember
	invitationDao           dao.Invitation
	messageDao              dao.Message
	activityCache           cache.Activity
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskService             service.Task
	teamService             service.Team
	sprintService           service.Sprint
	invitationService       service.Invitation
	userService             service.User
	stateSyncer             *realtime.StateSyncer
}

func NewDependencies(
	dataCollector obs.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
	userDao dao.User,
	teamDao dao.Team,
	taskDao dao.Task,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	messageDao dao.Message,
	activityCache cache.Activity,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskService service.Task,
	teamService service.Team,
	sprintService service.Sprint,
	userService service.User,
	invitationService service.Invitation,
	stateSyncer *realtime.StateSyncer,
) *Dependencies {
	return &Dependencies{
		dataCollector:           dataCollector,
		cloudClientRegistry:     cloudClientRegistry,
		userDao:                 userDao,
		teamDao:                 teamDao,
		taskDao:                 taskDao,
		teamMemberDao:           teamMemberDao,
		invitationDao:           invitationDao,
		messageDao:              messageDao,
		activityCache:           activityCache,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		taskService:             taskService,
		teamService:             teamService,
		sprintService:           sprintService,
		userService:             userService,
		invitationService:       invitationService,
		stateSyncer:             stateSyncer,
	}
}
