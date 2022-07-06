package service

import (
	"context"
	"log"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type CreateTaskInput struct {
	Goal        string
	Context     *string
	OwnerUserID *uint64
	DueAt       *time.Time
}

type Task struct {
	cloudClientRegistry *cloudAPI.ClientRegistry
	taskSyncer          collection.TaskSyncer
	threadService       Thread
}

func (t Task) CreateTask(ct context.Context, teamID uint64, taskInput CreateTaskInput) (entity.Task, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return entity.Task{}, err
	}

	genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
	genTaskIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	threadID, err := t.threadService.createThread(ct)
	if err != nil {
		return entity.Task{}, err
	}

	task := entity.Task{
		ID:               genTaskIDRes.UniqueNumber,
		Goal:             taskInput.Goal,
		Context:          taskInput.Context,
		Status:           entity.TaskStatusUpcoming,
		CreatorUserID:    userID,
		OwningTeamID:     teamID,
		OwnerUserID:      taskInput.OwnerUserID,
		CommentsThreadID: threadID,
		CreatedAt:        time.Now(),
		DueAt:            taskInput.DueAt,
	}

	err = t.taskSyncer.CreateAndSyncTask(task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func NewTask(
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskSyncer collection.TaskSyncer,
	threadService Thread,
) Task {
	return Task{
		cloudClientRegistry: cloudClientRegistry,
		taskSyncer:          taskSyncer,
		threadService:       threadService,
	}
}
