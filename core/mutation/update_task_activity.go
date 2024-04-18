package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/activity"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTaskActivity struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	activityCache    activity.Activity
	id               uint64
	taskActivity     entity.TaskActivity
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateTaskActivity)(nil)

func (u *UpdateTaskActivity) GetID() uint64 {
	return u.id
}

func (u *UpdateTaskActivity) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	_, err := u.activityCache.UpdateTaskActivity(ct, u.taskActivity.TeamID, u.taskActivity.TaskID, &u.taskActivity)
	if err != nil {
		return err
	}

	return nil
}

func (u *UpdateTaskActivity) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	var err *errs.Error
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.taskActivity.TeamID)
	if err != nil {
		return err
	}

	u.notifierPrepared = true
	return err
}

func (u *UpdateTaskActivity) Undo() *errs.Error {
	return nil
}

func (u *UpdateTaskActivity) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateTaskActivity) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TaskActivityCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.taskActivity,
	}
}

func (u *UpdateTaskActivity) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTaskActivity(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	activityCache activity.Activity,
	taskActivity entity.TaskActivity,
) *UpdateTaskActivity {
	return &UpdateTaskActivity{
		logger:           logger,
		stateSyncer:      stateSyncer,
		activityCache:    activityCache,
		id:               stateSyncer.NextMutationID(),
		taskActivity:     taskActivity,
		notifierPrepared: false,
	}
}
