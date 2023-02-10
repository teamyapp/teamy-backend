package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskLinkInput struct {
	TaskID       uint64
	Title        string
	URL          string
	IconURL      *string
	IconHoverURL *string
}

type TaskLink struct {
	dataCollector       telemetry.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	authorizer          Authorizer
	stateSyncer         *realtime.StateSyncer
	taskLinkDao         dao.TaskLink
	taskDao             dao.Task
}

func (t TaskLink) CreateTaskLink(ct context.Context, taskLinkEntity CreateTaskLinkInput) (entity.TaskLink, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.TaskLink{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateTaskLinkQuery(userID, taskLinkEntity.TaskID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.TaskLink{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.TaskLink{}, internalErr
		}
	}

	genTaskLinkIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskLinkID"}
	genTaskLinkIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskLinkIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.TaskLink{}, internalErr
	}

	taskLink := entity.TaskLink{
		ID:           genTaskLinkIDRes.UniqueNumber,
		TaskID:       taskLinkEntity.TaskID,
		Title:        taskLinkEntity.Title,
		URL:          taskLinkEntity.URL,
		IconURL:      taskLinkEntity.IconURL,
		IconHoverURL: taskLinkEntity.IconHoverURL,
		CreatedAt:    time.Now(),
	}
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	createTaskLinkMutation := mutation.NewCreateTaskLinkMutation(t.dataCollector, t.stateSyncer, t.taskLinkDao, t.taskDao, taskLink)

	err := realTimeTransaction.ApplyMutation(ct, createTaskLinkMutation)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TaskLink{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TaskLink{}, err
	}

	if feature.EnableAuthorization {
		err = t.authorizer.registerResource(ct, authorization.TaskLinkResourceType, taskLink.ID)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.TaskLink{}, err
		}

		err = t.authorizer.assignParentResource(ct, authorization.TaskLinkResourceType, taskLink.ID, authorization.TaskResourceType, taskLink.TaskID)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.TaskLink{}, err
		}
	}

	return taskLink, nil
}

func (t TaskLink) FindLinksByTaskID(ct context.Context, taskID uint64) ([]entity.TaskLink, *errs.Error) {
	return t.taskLinkDao.FindLinksByTaskID(ct, taskID)
}

func NewTaskLink(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	stateSyncer *realtime.StateSyncer,
	taskLinkDao dao.TaskLink,
	taskDao dao.Task,
) TaskLink {
	return TaskLink{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		authorizer:          authorizer,
		stateSyncer:         stateSyncer,
		taskLinkDao:         taskLinkDao,
		taskDao:             taskDao,
	}
}
