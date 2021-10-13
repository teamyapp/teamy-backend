package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Prioritization struct {
}

func (p Prioritization) prioritizeTasks(inputTasks []entity.Task) []entity.Task {
	return inputTasks
}

func (p Prioritization) SelectNeedAttention() *entity.Task {
	panic("not implemented")
}

func NewPrioritization() Prioritization {
	return Prioritization{}
}
