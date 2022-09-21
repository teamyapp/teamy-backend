package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type MessageSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	messageDao          dao.Message
	taskDao             dao.Task
}

func (m MessageSyncer) CreateAndSyncMessage(ct context.Context, message entity.Message) error {
	err := m.messageDao.CreateMessage(ct, message)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	task, err := m.taskDao.FindTaskByCommentsThreadID(ct, message.ThreadID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

func (m MessageSyncer) UpdateAndSyncMessage(ct context.Context, message entity.Message) error {
	err := m.messageDao.UpdateMessage(ct, message)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	task, err := m.taskDao.FindTaskByCommentsThreadID(ct, message.ThreadID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

func (m MessageSyncer) DeleteAndSyncMessage(ct context.Context, messageID uint64) error {
	message, err := m.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	task, err := m.taskDao.FindTaskByCommentsThreadID(ct, message.ThreadID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = m.messageDao.DeleteMessage(ct, messageID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task) MessageSyncer {
	return MessageSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		messageDao:          messageDao,
		taskDao:             taskDao,
	}
}
