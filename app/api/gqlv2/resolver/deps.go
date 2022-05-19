package resolver

import (
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/teamy-backend/app/dao"
)

type Dependencies struct {
	userDao        dao.User
	teamDao        dao.Team
	teamMemberDao  dao.TeamMember
	invitationDao  dao.Invitation
	taskDao        dao.Task
	messageDao     dao.Message
	cloudAPIClient *rpc.CloudAPIClient
}

func NewDependencies(
	userDao dao.User,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	taskDao dao.Task,
	messageDao dao.Message,
	cloudAPIClient *rpc.CloudAPIClient,
) *Dependencies {
	return &Dependencies{
		userDao:        userDao,
		teamDao:        teamDao,
		teamMemberDao:  teamMemberDao,
		invitationDao:  invitationDao,
		taskDao:        taskDao,
		messageDao:     messageDao,
		cloudAPIClient: cloudAPIClient,
	}
}
