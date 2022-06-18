package collection

import (
	"strconv"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const taskCollectionType = "Task"

type TaskSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	taskDao            dao.Task
}

func (t TaskSyncer) CreateAndSyncTask(task entity.Task) error {
	err := t.taskDao.CreateTask(task)
	if err != nil {
		return err
	}

	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: taskCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes:     storage.MapAttributes(task),
	})
}

func (t TaskSyncer) UpdateAndSyncTask(task entity.Task) error {
	err := t.taskDao.UpdateTask(task)
	if err != nil {
		return err
	}

	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: taskCollectionType,
		MutationType:   storage.UpdateMutationType,
		Attributes:     storage.MapAttributes(task),
	})
}

func (t TaskSyncer) DeleteAndSyncTask(taskID uint64) error {
	err := t.taskDao.DeleteTask(taskID)
	if err != nil {
		return err
	}

	idStr := strconv.FormatUint(taskID, 10)
	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: taskCollectionType,
		MutationType:   storage.DeleteMutationType,
		Attributes: map[string]*string{
			"ID": &idStr,
		},
	})
}

func NewTaskSyncer(realTimeCollection *storage.RealTimeCollections, taskDao dao.Task) TaskSyncer {
	realTimeCollection.RegisterCollectionType(taskCollectionType)
	return TaskSyncer{
		realTimeCollection: realTimeCollection,
		taskDao:            taskDao,
	}
}
