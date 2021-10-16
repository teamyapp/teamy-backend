package resolver

import (
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/service"
)

type ExecutionMode struct {
	userID           oneEntity.ID
	executionService service.Execution
}

func (e ExecutionMode) CurrUserStatus() PersonalStatus {
	panic("not implemented")
}

func (e ExecutionMode) TeamStatus() (TeamStatus, error) {
	teamStatus, err := e.executionService.GetActiveTeamStatus(e.userID)
	if err != nil {
		log.Println(err)
	}
	return TeamStatus{
		teamStatus: teamStatus,
	}, err
}
