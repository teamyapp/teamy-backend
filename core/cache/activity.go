package cache

import (
	"errors"

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

func (a Activity) FindAllTaskActivitiesByTeamID(teamID uint64) (map[uint64]*entity.TaskActivity, error) {
	teamActivity := a.GetOrInitTeamActivity(teamID)
	return teamActivity.TaskActivities, nil
}

func (a Activity) UpdateTaskActivity(teamID uint64, taskID uint64, taskActivity *entity.TaskActivity) (*entity.TaskActivity, error) {
	teamActivity, ok := a.teamActivities[teamID]

	if !ok {
		err := errors.New("teamActivity not found")
		a.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"teamID": teamID,
			},
		})
		return nil, err
	}

	teamActivity.TaskActivities[taskID] = taskActivity
	return taskActivity, nil
}

func NewActivity(dataCollector obs.DataCollector) Activity {
	return Activity{
		dataCollector:  dataCollector,
		teamActivities: map[uint64]*entity.TeamActivity{},
	}
}
