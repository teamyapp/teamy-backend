package collection

import (
	"strconv"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const threadCollectionType = "Thread"

type ThreadSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	threadDao          dao.Thread
}

func (t ThreadSyncer) CreateAndSyncThread(threadID uint64) error {
	err := t.threadDao.CreateThread(threadID)
	if err != nil {
		return err
	}

	idStr := strconv.FormatUint(threadID, 10)
	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: threadCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes: map[string]*string{
			"ID": &idStr,
		},
	})
}

func (t ThreadSyncer) DeleteAndSyncThread(threadID uint64) error {
	err := t.threadDao.DeleteThread(threadID)
	if err != nil {
		return err
	}

	idStr := strconv.FormatUint(threadID, 10)
	return t.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: threadCollectionType,
		MutationType:   storage.DeleteMutationType,
		Attributes: map[string]*string{
			"ID": &idStr,
		},
	})
}

func NewThreadSyncer(realTimeCollection *storage.RealTimeCollections, threadDao dao.Thread) ThreadSyncer {
	realTimeCollection.RegisterCollectionType(threadCollectionType)
	return ThreadSyncer{
		realTimeCollection: realTimeCollection,
		threadDao:          threadDao,
	}
}
