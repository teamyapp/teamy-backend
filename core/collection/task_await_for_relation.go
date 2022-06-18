package collection

import (
	"strconv"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const taskAwaitForRelationCollectionType = "TaskAwaitForRelation"

type TaskAwaitForRelationSyncer struct {
	realTimeCollection      *storage.RealTimeCollections
	taskAwaitForRelationDao dao.TaskAwaitForRelation
}

func (t TaskAwaitForRelationSyncer) CreateAndSyncRelation(relation entity.TaskAwaitForRelation) error {
	err := t.taskAwaitForRelationDao.CreateRelation(relation)
	if err != nil {
		return err
	}

	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: taskAwaitForRelationCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes:     storage.MapAttributes(relation),
	})
}

func (t TaskAwaitForRelationSyncer) DeleteAndSyncRelation(waitingTaskID uint64, awaitForTaskID uint64) error {
	err := t.taskAwaitForRelationDao.DeleteRelation(waitingTaskID, awaitForTaskID)
	if err != nil {
		return err
	}

	waitingTaskIDStr := strconv.FormatUint(waitingTaskID, 10)
	awaitTaskUserIDStr := strconv.FormatUint(awaitForTaskID, 10)
	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: taskAwaitForRelationCollectionType,
		MutationType:   storage.DeleteMutationType,
		Attributes: map[string]*string{
			"AWaitingTaskID": &waitingTaskIDStr,
			"AWaitForTaskID": &awaitTaskUserIDStr,
		},
	})
}

func NewTaskAwaitForRelationSyncer(
	realTimeCollection *storage.RealTimeCollections,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
) TaskAwaitForRelationSyncer {
	realTimeCollection.RegisterCollectionType(taskAwaitForRelationCollectionType)
	return TaskAwaitForRelationSyncer{
		realTimeCollection:      realTimeCollection,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
	}
}
