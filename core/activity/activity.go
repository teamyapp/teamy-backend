package activity

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activity struct {
	logger         telemetry.Logger
	teamActivities map[uint64]*entity.TeamActivity
}

func (a Activity) GetTeamActivity(teamID uint64) *entity.TeamActivity {
	teamActivity, ok := a.teamActivities[teamID]
	if !ok {
		return nil
	}

	return teamActivity
}

func (a Activity) InitTeamActivity(teamID uint64) *entity.TeamActivity {
	teamActivity := &entity.TeamActivity{}
	teamActivity.TaskActivities = map[uint64]*entity.TaskActivity{}
	a.teamActivities[teamID] = teamActivity
	return teamActivity
}

func (a Activity) GetOrInitTeamActivity(teamID uint64) *entity.TeamActivity {
	teamActivity := a.GetTeamActivity(teamID)
	if teamActivity == nil {
		return a.InitTeamActivity(teamID)
	}

	return teamActivity
}

func (a Activity) FindAllTaskActivitiesByTeamID(teamID uint64) map[uint64]*entity.TaskActivity {
	teamActivity := a.GetOrInitTeamActivity(teamID)
	return teamActivity.TaskActivities
}

func (a Activity) UpdateTaskActivity(ct context.Context, teamID uint64, taskID uint64, taskActivity *entity.TaskActivity) (*entity.TaskActivity, *errs.Error) {
	teamActivity, ok := a.teamActivities[teamID]
	if !ok {
		err := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("teamActivity not found: teamID=%v", teamID),
		}
		a.logger.ErrorWithContext(ct, err)
		return nil, err
	}

	teamActivity.TaskActivities[taskID] = taskActivity
	return taskActivity, nil
}

func NewActivity(logger telemetry.Logger) Activity {
	return Activity{
		logger:         logger,
		teamActivities: map[uint64]*entity.TeamActivity{},
	}
}
