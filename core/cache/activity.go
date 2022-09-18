package cache

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activity struct {
	dataCollector  obs.DataCollector
	teamActivities map[uint64]*entity.TeamActivity
}

func (a Activity) GetTeamActivity(teamID uint64) *entity.TeamActivity {
	teamActivity, ok := a.teamActivities[teamID]
	if !ok {
		return nil
	}

	return teamActivity
}

func (a Activity) AddTeamActivity(teamID uint64) *entity.TeamActivity {
	teamActivity := &entity.TeamActivity{}
	teamActivity.TaskActivities = map[uint64]*entity.TaskActivity{}
	a.teamActivities[teamID] = teamActivity

	return teamActivity
}

func (a Activity) GetOrAddTeamActivity(teamID uint64) *entity.TeamActivity {
	teamActivity := a.GetTeamActivity(teamID)

	if teamActivity == nil {
		return a.AddTeamActivity(teamID)
	}

	return teamActivity
}

func (a Activity) FindAllTaskActivitiesByTeamID(teamID uint64) ([]entity.TaskActivity, error) {
	teamActivity := a.GetOrAddTeamActivity(teamID)
	return teamActivity.TaskActivities, nil
}

func (a Activity) UpdateTaskActivity(taskID uint64, teamActivity *entity.TeamActivity, taskActivity *entity.TaskActivity) *entity.TaskActivity {
	teamActivity.TaskActivities[taskID] = taskActivity
	return taskActivity
}

func NewActivity(dataCollector obs.DataCollector) Activity {
	return Activity{dataCollector: dataCollector, teamActivities: map[uint64]*entity.TeamActivity{}}
}
