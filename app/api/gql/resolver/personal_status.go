package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type PersonalStatus struct {
	deps           *Dependencies
	personalStatus entity.PersonalStatus
}

func (p PersonalStatus) TaskNeedAttention() *Task {
	if p.personalStatus.TaskNeedAttention == nil {
		return nil
	}

	task := newTask(p.deps, *p.personalStatus.TaskNeedAttention)
	return &task
}

func (p PersonalStatus) UpcomingTasks() []Task {
	return toGraphQLTasks(p.deps, p.personalStatus.UpcomingTasks)
}

func (p PersonalStatus) DeliveredTasks() []Task {
	return toGraphQLTasks(p.deps, p.personalStatus.DeliveredTasks)
}
