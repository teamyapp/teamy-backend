package resolver

import "github.com/teamyapp/teamy-backend/app/service"

type Dependencies struct {
	userService      service.User
	taskService      service.Task
	teamService      service.Team
	executionService service.Execution
}

func NewDependencies(
	userService service.User,
	taskService service.Task,
	teamService service.Team,
	executionService service.Execution,
) *Dependencies {
	return &Dependencies{
		userService:      userService,
		taskService:      taskService,
		teamService:      teamService,
		executionService: executionService,
	}
}
