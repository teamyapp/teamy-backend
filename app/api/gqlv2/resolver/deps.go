package resolver

import (
	"github.com/teamyapp/teamy-backend/app/dao"
)

type Dependencies struct {
	userDao       dao.User
	teamDao       dao.Team
	teamMemberDao dao.TeamMember
	invitationDao dao.Invitation
	taskDao       dao.Task
	messageDao    dao.Message
}

func NewDependencies(
	userDao dao.User,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	taskDao dao.Task,
	messageDao dao.Message,
) *Dependencies {
	return &Dependencies{
		userDao:       userDao,
		teamDao:       teamDao,
		teamMemberDao: teamMemberDao,
		invitationDao: invitationDao,
		taskDao:       taskDao,
		messageDao:    messageDao,
	}
}
