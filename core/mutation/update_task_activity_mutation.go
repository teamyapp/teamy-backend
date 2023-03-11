package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTaskActivityMutation struct {
	dataCollector    telemetry.DataCollector
	stateSyncer      *realtime.StateSyncer
	activityCache    cache.Activity
	id               uint64
	taskActivity     entity.TaskActivity
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateTaskActivityMutation)(nil)

func (u *UpdateTaskActivityMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateTaskActivityMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	_, err := u.activityCache.UpdateTaskActivity(ct, u.taskActivity.TeamID, u.taskActivity.TaskID, &u.taskActivity)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateTaskActivityMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	var err *errs.Error
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.taskActivity.TeamID)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	u.notifierPrepared = true
	return err
}

func (u *UpdateTaskActivityMutation) Execute(ct context.Context) *errs.Error {
	_, err := u.activityCache.UpdateTaskActivity(ct, u.taskActivity.TeamID, u.taskActivity.TaskID, &u.taskActivity)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateTaskActivityMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateTaskActivityMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.taskActivity.TeamID)
}

func (u *UpdateTaskActivityMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateTaskActivityMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TaskActivityCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.taskActivity,
	}
}

func (u *UpdateTaskActivityMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTaskActivityMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	activityCache cache.Activity,
	taskActivity entity.TaskActivity,
) *UpdateTaskActivityMutation {
	return &UpdateTaskActivityMutation{
		dataCollector:    dataCollector,
		stateSyncer:      stateSyncer,
		activityCache:    activityCache,
		id:               stateSyncer.NextMutationID(),
		taskActivity:     taskActivity,
		notifierPrepared: false,
	}
}
