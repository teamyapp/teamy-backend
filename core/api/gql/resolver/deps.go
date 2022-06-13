package resolver

import (
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Dependencies struct {
	userDao                 dao.User
	teamDao                 dao.Team
	teamMemberDao           dao.TeamMember
	invitationDao           dao.Invitation
	taskDao                 dao.Task
	threadDao               dao.Thread
	messageDao              dao.Message
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	cloudAPIClient          *rpc.CloudAPIClient
}

func NewDependencies(
	userDao dao.User,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	taskDao dao.Task,
	threadDao dao.Thread,
	messageDao dao.Message,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	cloudAPIClient *rpc.CloudAPIClient,
) *Dependencies {
	return &Dependencies{
		userDao:                 userDao,
		teamDao:                 teamDao,
		teamMemberDao:           teamMemberDao,
		invitationDao:           invitationDao,
		taskDao:                 taskDao,
		threadDao:               threadDao,
		messageDao:              messageDao,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		cloudAPIClient:          cloudAPIClient,
	}
}
