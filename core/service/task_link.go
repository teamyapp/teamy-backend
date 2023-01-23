package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
)

type CreateTaskLinkInput struct {
	TaskID  uint64
	Title   string
	Url     string
	IconUrl *string
}

type TaskLink struct {
	dataCollector       obs.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	authorizer          Authorizer
	taskLinkDao         dao.TaskLink
}

func (t TaskLink) CreateTaskLink(ct context.Context, taskLinkEntity CreateTaskLinkInput) (entity.TaskLink, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TaskLink{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateTaskLinkQuery(userID, taskLinkEntity.TaskID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TaskLink{}, err
		}

		if !hasPermission {
			return entity.TaskLink{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	genTaskLinkIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskLinkID"}
	genTaskLinkIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskLinkIDReq)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TaskLink{}, err
	}

	taskLink := entity.TaskLink{
		ID:        genTaskLinkIDRes.UniqueNumber,
		TaskID:    taskLinkEntity.TaskID,
		Title:     taskLinkEntity.Title,
		Url:       taskLinkEntity.Url,
		IconUrl:   taskLinkEntity.IconUrl,
		CreatedAt: time.Now(),
	}

	err = t.taskLinkDao.CreateTaskLink(ct, taskLink)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TaskLink{}, err
	}

	if feature.EnableAuthorization {
		err = t.authorizer.registerResource(ct, authorization.TaskLinkResourceType, taskLink.ID)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TaskLink{}, err
		}

		err = t.authorizer.assignParentResource(ct, authorization.TaskLinkResourceType, taskLink.ID, authorization.TaskResourceType, taskLink.TaskID)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TaskLink{}, err
		}
	}

	return taskLink, nil
}

func (t TaskLink) FindLinksByTaskID(ct context.Context, taskID uint64) ([]entity.TaskLink, error) {
	return t.taskLinkDao.FindLinksByTaskID(ct, taskID)
}

func NewTaskLink(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	taskLinkDao dao.TaskLink) TaskLink {
	return TaskLink{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		authorizer:          authorizer,
		taskLinkDao:         taskLinkDao,
	}
}
