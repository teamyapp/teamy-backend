package resolver

import (
	"github.com/teamyapp/teamy-backend/app/dao"
)

type Dependencies struct {
	userDao dao.User
	teamDao dao.Team
}

func NewDependencies(
	userDao dao.User,
	teamDao dao.Team,
) Dependencies {
	return Dependencies{
		userDao: userDao,
		teamDao: teamDao,
	}
}
