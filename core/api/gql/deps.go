package gql

import (
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Dependencies struct {
	dataCollector              obs.DataCollector
	cloudClientRegistry        *api.ClientRegistry
	userDao                    dao.User
	teamDao                    dao.Team
	teamMemberDao              dao.TeamMember
	invitationDao              dao.Invitation
	messageDao                 dao.Message
	taskAwaitForRelationDao    dao.TaskAwaitForRelation
	userSyncer                 collection.UserSyncer
	teamSyncer                 collection.TeamSyncer
	teamMemberSyncer           collection.TeamMemberSyncer
	invitationSyncer           collection.InvitationSyncer
	messageSyncer              collection.MessageSyncer
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer
	taskService                service.Task
	teamService                service.Team
	sprintService              service.Sprint
	userService                service.User
}

func NewDependencies(
	dataCollector obs.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
	userDao dao.User,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	messageDao dao.Message,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	userSyncer collection.UserSyncer,
	teamSyncer collection.TeamSyncer,
	teamMemberSyncer collection.TeamMemberSyncer,
	invitationSyncer collection.InvitationSyncer,
	messageSyncer collection.MessageSyncer,
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer,
	taskService service.Task,
	teamService service.Team,
	sprintService service.Sprint,
	userService service.User,
) *Dependencies {
	return &Dependencies{
		dataCollector:              dataCollector,
		cloudClientRegistry:        cloudClientRegistry,
		userDao:                    userDao,
		teamDao:                    teamDao,
		teamMemberDao:              teamMemberDao,
		invitationDao:              invitationDao,
		messageDao:                 messageDao,
		taskAwaitForRelationDao:    taskAwaitForRelationDao,
		userSyncer:                 userSyncer,
		teamSyncer:                 teamSyncer,
		teamMemberSyncer:           teamMemberSyncer,
		invitationSyncer:           invitationSyncer,
		messageSyncer:              messageSyncer,
		taskAwaitForRelationSyncer: taskAwaitForRelationSyncer,
		taskService:                taskService,
		teamService:                teamService,
		sprintService:              sprintService,
		userService:                userService,
	}
}
