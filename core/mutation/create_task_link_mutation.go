package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskLinkMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	taskLinkDao   dao.TaskLink
	taskDao       dao.Task
	id            uint64
	taskLink      entity.TaskLink
}

func (c *CreateTaskLinkMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskLinkMutation) Execute(ct context.Context) *errs.Error {
	err := c.taskLinkDao.CreateTaskLink(ct, c.taskLink)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateTaskLinkMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTaskLinkMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	task, err := c.taskDao.FindTaskByID(ct, c.taskLink.TaskID)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (c *CreateTaskLinkMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskLinkCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.taskLink,
	}
}

func (c *CreateTaskLinkMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTaskLinkMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskLinkDao dao.TaskLink,
	taskDao dao.Task,
	taskLink entity.TaskLink,
) *CreateTaskLinkMutation {
	return &CreateTaskLinkMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		taskLinkDao:   taskLinkDao,
		taskDao:       taskDao,
		id:            stateSyncer.NextMutationID(),
		taskLink:      taskLink,
	}
}
