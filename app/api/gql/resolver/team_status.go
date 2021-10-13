package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type TeamStatus struct {
	teamStatus entity.TeamStatus
}

func (t TeamStatus) UpcomingTasks() []Task {
	return toGraphQLTasks(t.teamStatus.UpcomingTasks)
}

func (t TeamStatus) InProgressTasks() []Task {
	return toGraphQLTasks(t.teamStatus.InProgressTasks)
}

func (t TeamStatus) DeliveredTasks() []Task {
	return toGraphQLTasks(t.teamStatus.DeliveredTasks)
}
