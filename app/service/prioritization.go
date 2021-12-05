package service

import (
	"sort"

	"github.com/teamyapp/teamy-backend/app/entity"
)

type Prioritization struct {
}

func (p Prioritization) PrioritizeTasks(inputTasks []entity.Task) []entity.Task {
	sort.SliceStable(inputTasks, func(i, j int) bool {
		if inputTasks[i].DueAt == nil && inputTasks[j].DueAt == nil {
			return inputTasks[i].ID < inputTasks[j].ID
		} else if inputTasks[i].DueAt == nil {
			return false
		} else if inputTasks[j].DueAt == nil {
			return true
		}
		return inputTasks[i].DueAt.Before(*inputTasks[j].DueAt)
	})

	return inputTasks
}

func NewPrioritization() Prioritization {
	return Prioritization{}
}
