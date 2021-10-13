package resolver

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/service"
)

type ExecutionMode struct {
	userID           entity.ID
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
