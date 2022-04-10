package resolver

import (
	"github.com/teamyapp/teamy-backend/app/dao"
)

type Dependencies struct {
	userDao       dao.User
	teamDao       dao.Team
	teamMemberDao dao.TeamMember
}

func NewDependencies(
	userDao dao.User,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
) Dependencies {
	return Dependencies{
		userDao:       userDao,
		teamDao:       teamDao,
		teamMemberDao: teamMemberDao,
	}
}
