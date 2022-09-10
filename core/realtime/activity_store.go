package realtime

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ActivityStore struct {
	userActivities  map[uint64]*UserActivity
	clientNotifiers map[uint64]*ClientNotifier
	dataCollector   obs.DataCollector
}

func (a *ActivityStore) registerClientNotifier(userID uint64, clientID uint64, clientNotifier *ClientNotifier) {
	a.clientNotifiers[clientID] = clientNotifier
	go func() {
		<-clientNotifier.subscribeClientDisconnect()
		a.dataCollector.Logger.Log(obs.Info, obs.Props{
			obs.MessageProp: obs.Props{
				"summary":  "client disconnected",
				"clientID": clientID,
			},
		})
		a.unregisterClientNotifier(userID, clientID)
	}()
}

func (a *ActivityStore) unregisterClientNotifier(userID uint64, clientID uint64) {
	delete(a.clientNotifiers, clientID)
	for _, teamActivity := range a.userActivities[userID].teamActivities {
		if teamActivity.clientActivities[clientID] != nil {
			teamActivity.clientActivities[clientID] = nil
		}

		if len(teamActivity.clientActivities) == 0 {
			delete(a.userActivities[userID].teamActivities, teamActivity.teamID)
		}

		if len(a.userActivities[userID].teamActivities) == 0 {
			a.userActivities[userID].state = entity.UserStatusDisConnected
		}
	}
}

func (a *ActivityStore) getUserActivity(userID uint64) *UserActivity {
	userActivity, ok := a.userActivities[userID]
	if !ok {
		userActivity = a.newUserActivity(userID)
	}

	userActivity.state = entity.UserStatusConnected
	return userActivity
}

func (a *ActivityStore) StartDraggingTask(userID uint64, clientID uint64, taskID uint64, teamID uint64) {
	userActivity := a.getUserActivity(userID)

	teamActivity := userActivity.getTeamActivity(teamID)
	clientActivity := teamActivity.getClientActivity(clientID)
	clientActivity.taskAction.taskID = taskID
	clientActivity.taskAction.dragging = true
}

func (a *ActivityStore) StopDraggingTask(userID uint64, clientID uint64, taskID uint64, teamID uint64) {
	userActivity := a.getUserActivity(userID)

	teamActivity := userActivity.getTeamActivity(teamID)
	clientActivity := teamActivity.getClientActivity(clientID)
	clientActivity.taskAction.taskID = taskID
	clientActivity.taskAction.dragging = false
}

func (a *ActivityStore) newUserActivity(userID uint64) *UserActivity {
	userActivity := newUserActivity(userID)
	a.userActivities[userID] = userActivity
	return userActivity
}

func NewActivityStore(dataCollector obs.DataCollector) *ActivityStore {
	return &ActivityStore{
		dataCollector:   dataCollector,
		userActivities:  map[uint64]*UserActivity{},
		clientNotifiers: map[uint64]*ClientNotifier{},
	}
}
