package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type MessageSyncer struct {
	realTimeStateSyncer *realtime.StateSyncer
	messageDao          dao.Message
	taskDao             dao.Task
}

func (m MessageSyncer) CreateAndSyncMessage(message entity.Message) error {
	err := m.messageDao.CreateMessage(message)
	if err != nil {
		return err
	}

	task, err := m.taskDao.FindTaskByCommentThreadID(message.ThreadID)
	if err != nil {
		return err
	}

	m.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: message,
	})
	return nil
}

func (m MessageSyncer) UpdateAndSyncMessage(message entity.Message) error {
	err := m.messageDao.UpdateMessage(message)
	if err != nil {
		return err
	}

	task, err := m.taskDao.FindTaskByCommentThreadID(message.ThreadID)
	if err != nil {
		return err
	}

	m.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: message,
	})
	return nil
}

func (m MessageSyncer) DeleteAndSyncMessage(messageID uint64) error {
	message, err := m.messageDao.FindMessageByID(messageID)
	if err != nil {
		return err
	}

	task, err := m.taskDao.FindTaskByCommentThreadID(message.ThreadID)
	if err != nil {
		return err
	}

	err = m.messageDao.DeleteMessage(messageID)
	if err != nil {
		return err
	}

	m.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: messageID,
	})
	return nil
}

func NewMessageSyncer(
	realTimeStateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task) MessageSyncer {
	return MessageSyncer{
		realTimeStateSyncer: realTimeStateSyncer,
		messageDao:          messageDao,
		taskDao:             taskDao,
	}
}
