package resolver

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"

	oneEntity "github.com/teamyapp/one/entity"
)

type ExecutionMode struct {
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	userID        oneEntity.ID
}

func (e ExecutionMode) PersonalStatus() (PersonalStatus, error) {
	personalStatus, err := e.deps.executionService.GetPersonalStatusForActiveTeam(e.userID)
	if err != nil {
		log.Println(err)
	}
	return PersonalStatus{personalStatus: personalStatus}, err
}

func (e ExecutionMode) TeamStatus() (TeamStatus, error) {
	teamStatus, err := e.deps.executionService.GetActiveTeamStatus(e.userID)
	if err != nil {
		log.Println(err)
	}
	return TeamStatus{
		teamStatus: teamStatus,
	}, err
}

func newExecutionMode(deps *Dependencies, prototypeDeps *resolver.Dependencies, userID oneEntity.ID) ExecutionMode {
	return ExecutionMode{
		deps:          deps,
		prototypeDeps: prototypeDeps,
		userID:        userID,
	}
}
