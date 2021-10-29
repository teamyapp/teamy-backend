package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type PersonalStatus struct {
	personalStatus entity.PersonalStatus
}

func (p PersonalStatus) NeedAttention() *Task {
	if p.personalStatus.TaskNeedAttention == nil {
		return nil
	}

	task := newTask(*p.personalStatus.TaskNeedAttention)
	return &task
}

func (p PersonalStatus) UpcomingTasks() []Task {
	return toGraphQLTasks(p.personalStatus.UpcomingTasks)
}

func (p PersonalStatus) DeliveredTasks() []Task {
	return toGraphQLTasks(p.personalStatus.DeliveredTasks)
}
