package resolver

import (
	"github.com/teamyapp/teamy-backend/app/api/gql/datastore"
	"github.com/teamyapp/teamy-backend/app/repo"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Dependencies struct {
	// taskRepo              repo.Task
	teamRepo              repo.Team
	prioritizationService service.Prioritization
	Data                  *datastore.DataStore
}

func NewDependencies(
	// taskRepo repo.Task,
	teamRepo repo.Team,
	prioritizationService service.Prioritization,
	data *datastore.DataStore,
) *Dependencies {
	return &Dependencies{
		// taskRepo:              taskRepo,
		teamRepo:              teamRepo,
		prioritizationService: prioritizationService,
		Data:                  data,
	}
}
