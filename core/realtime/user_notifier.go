package realtime

import (
	"github.com/teamyapp/cloud/libs/telemetry"
)

type UserNotifier struct {
	dataCollector             telemetry.DataCollector
	userID                    uint64
	userDisconnectCh          chan bool
	userDisconnectSubscribers []chan bool
	clientNotifiers           map[uint64]*ClientNotifier
}

func (u *UserNotifier) GetClientNotifiers() map[uint64]*ClientNotifier {
	return u.clientNotifiers
}

func (u *UserNotifier) subscribeUserDisconnect() chan bool {
	subscriber := make(chan bool)
	u.userDisconnectSubscribers = append(u.userDisconnectSubscribers, subscriber)
	return subscriber
}

func (u UserNotifier) registerClientNotifier(clientID uint64, clientNotifier *ClientNotifier) {
	u.clientNotifiers[clientID] = clientNotifier
	go func() {
		<-clientNotifier.subscribeClientDisconnect()
		u.unregisterClientNotifier(clientID)
	}()
}

func (u UserNotifier) unregisterClientNotifier(clientID uint64) {
	delete(u.clientNotifiers, clientID)
	if len(u.clientNotifiers) == 0 {
		u.userDisconnectCh <- true
	}
}

func newUserNotifier(dataCollector telemetry.DataCollector, userID uint64) *UserNotifier {
	userNotifier := &UserNotifier{
		dataCollector:             dataCollector,
		userID:                    userID,
		clientNotifiers:           map[uint64]*ClientNotifier{},
		userDisconnectSubscribers: make([]chan bool, 0),
		userDisconnectCh:          make(chan bool),
	}
	go func() {
		<-userNotifier.userDisconnectCh
		for _, subscriber := range userNotifier.userDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return userNotifier
}
