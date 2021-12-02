package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Prioritization struct {
}

func (p Prioritization) PrioritizeTasks(inputTasks []entity.Task) []entity.Task {
	return inputTasks
}

func NewPrioritization() Prioritization {
	return Prioritization{}
}
