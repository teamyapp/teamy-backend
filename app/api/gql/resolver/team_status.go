package resolver

import (
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type TeamStatus struct {
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	teamStatus    entity.TeamStatus
}

func (t TeamStatus) UpcomingTasks() []Task {
	return toGraphQLTasks(t.deps, t.prototypeDeps, t.teamStatus.UpcomingTasks)
}

func (t TeamStatus) InProgressTasks() []Task {
	return toGraphQLTasks(t.deps, t.prototypeDeps, t.teamStatus.InProgressTasks)
}

func (t TeamStatus) DeliveredTasks() []Task {
	return toGraphQLTasks(t.deps, t.prototypeDeps, t.teamStatus.DeliveredTasks)
}
