package resolver

import (
	"github.com/teamyapp/teamy-backend/app/repo"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Dependencies struct {
	userRepo              repo.User
	taskRepo              repo.Task
	teamRepo              repo.Team
	prioritizationService service.Prioritization
}

func NewDependencies(
	userRepo repo.User,
	taskRepo repo.Task,
	teamRepo repo.Team,
	prioritizationService service.Prioritization,
) *Dependencies {
	return &Dependencies{
		userRepo:              userRepo,
		taskRepo:              taskRepo,
		teamRepo:              teamRepo,
		prioritizationService: prioritizationService,
	}
}
