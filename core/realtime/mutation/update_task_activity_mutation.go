package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTaskActivityMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	activityCache cache.Activity
	taskActivity  entity.TaskActivity
	dataCollector obs.DataCollector
}

func (c *UpdateTaskActivityMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateTaskActivityMutation) Execute(ct context.Context) error {
	_, err := u.activityCache.UpdateTaskActivity(ct, u.taskActivity.TeamID, u.taskActivity.TaskID, &u.taskActivity)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateTaskActivityMutation) Undo() error {
	return nil
}

func (u *UpdateTaskActivityMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamID)
}

func (u *UpdateTaskActivityMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TaskActivityCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.taskActivity,
	}
}

func NewUpdateTaskActivityMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	activityCache cache.Activity,
	taskActivity entity.TaskActivity,
	dataCollector obs.DataCollector) *UpdateTaskActivityMutation {
	return &UpdateTaskActivityMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		activityCache: activityCache,
		taskActivity:  taskActivity,
		dataCollector: dataCollector,
	}
}
