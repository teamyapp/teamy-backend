package resolver

import (
	"github.com/teamyapp/teamy-backend/app/api/gql/datastore"
	"github.com/teamyapp/teamy-backend/app/repo"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Dependencies struct {
	userRepo              repo.User
	taskRepo              repo.Task
	teamRepo              repo.Team
	prioritizationService service.Prioritization
	Data                  *datastore.DataStore
}

func NewDependencies(
	userRepo repo.User,
	taskRepo repo.Task,
	teamRepo repo.Team,
	prioritizationService service.Prioritization,
	data *datastore.DataStore,
) *Dependencies {
	return &Dependencies{
		userRepo:              userRepo,
		taskRepo:              taskRepo,
		teamRepo:              teamRepo,
		prioritizationService: prioritizationService,
		Data:                  data,
	}
}
