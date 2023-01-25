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
	stateSyncer             *realtime.StateSyncer
	userDao                 dao.User
	teamDao                 dao.Team
	taskDao                 dao.Task
	teamMemberDao           dao.TeamMember
	invitationDao           dao.Invitation
	messageDao              dao.Message
	activityCache           cache.Activity
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskService             service.Task
	taskLinkService         service.TaskLink
	teamService             service.Team
	sprintService           service.Sprint
	invitationService       service.Invitation
	userService             service.User
}

func NewDependencies(
	dataCollector obs.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	userDao dao.User,
	teamDao dao.Team,
	taskDao dao.Task,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	messageDao dao.Message,
	activityCache cache.Activity,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskService service.Task,
	taskLinkService service.TaskLink,
	teamService service.Team,
	sprintService service.Sprint,
	userService service.User,
	invitationService service.Invitation,
) *Dependencies {
	return &Dependencies{
		dataCollector:           dataCollector,
		cloudClientRegistry:     cloudClientRegistry,
		stateSyncer:             stateSyncer,
		userDao:                 userDao,
		teamDao:                 teamDao,
		taskDao:                 taskDao,
		teamMemberDao:           teamMemberDao,
		invitationDao:           invitationDao,
		messageDao:              messageDao,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		activityCache:           activityCache,
		taskService:             taskService,
		taskLinkService:         taskLinkService,
		teamService:             teamService,
		sprintService:           sprintService,
		userService:             userService,
		invitationService:       invitationService,
	}
}
