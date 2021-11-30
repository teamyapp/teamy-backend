package resolver

import (
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type PersonalStatus struct {
	deps           *Dependencies
	prototypeDeps  *resolver.Dependencies
	personalStatus entity.PersonalStatus
}

func (p PersonalStatus) TaskNeedAttention() *Task {
	if p.personalStatus.TaskNeedAttention == nil {
		return nil
	}

	task := newTask(p.deps, p.prototypeDeps, *p.personalStatus.TaskNeedAttention)
	return &task
}

func (p PersonalStatus) UpcomingTasks() []Task {
	return toGraphQLTasks(p.deps, p.prototypeDeps, p.personalStatus.UpcomingTasks)
}

func (p PersonalStatus) DeliveredTasks() []Task {
	return toGraphQLTasks(p.deps, p.prototypeDeps, p.personalStatus.DeliveredTasks)
}
