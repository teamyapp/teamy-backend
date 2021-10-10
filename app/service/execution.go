package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Execution struct {
}

func (e Execution) GetPersonalStatus(userID int) (entity.PersonalStatus, error) {
	panic("not implemented")
}

func (e Execution) GetActiveTeamStatus() (entity.TeamStatus, error) {
	panic("not implemented")
}
