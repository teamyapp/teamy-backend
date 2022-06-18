package collection

import (
	"strconv"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const messageCollectionType = "Message"

type MessageSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	messageDao         dao.Message
}

func (m MessageSyncer) CreateAndSyncMessage(message entity.Message) error {
	err := m.messageDao.CreateMessage(message)
	if err != nil {
		return err
	}

	return m.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: messageCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes:     storage.MapAttributes(message),
	})
}

func (m MessageSyncer) UpdateAndSyncMessage(message entity.Message) error {
	err := m.messageDao.UpdateMessage(message)
	if err != nil {
		return err
	}

	return m.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: messageCollectionType,
		MutationType:   storage.UpdateMutationType,
		Attributes:     storage.MapAttributes(message),
	})
}

func (m MessageSyncer) DeleteAndSyncMessage(messageID uint64) error {
	err := m.messageDao.DeleteMessage(messageID)
	if err != nil {
		return err
	}

	idStr := strconv.FormatUint(messageID, 10)
	return m.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: messageCollectionType,
		MutationType:   storage.DeleteMutationType,
		Attributes: map[string]*string{
			"ID": &idStr,
		},
	})
}

func NewMessageSyncer(realTimeCollection *storage.RealTimeCollections, messageDao dao.Message) MessageSyncer {
	realTimeCollection.RegisterCollectionType(messageCollectionType)
	return MessageSyncer{
		realTimeCollection: realTimeCollection,
		messageDao:         messageDao,
	}
}
