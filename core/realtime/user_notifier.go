package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
)

type UserNotifier struct {
	dataCollector             obs.DataCollector
	userID                    uint64
	userDisconnectCh          chan bool
	userDisconnectSubscribers []chan bool
	clientNotifiers           map[uint64]*ClientNotifier
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

func (u UserNotifier) notifyTransaction(ct context.Context, transaction *Transaction) {
	u.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"Summary": "notify transaction",
			"UserId":  u.userID,
		},
	})
	for _, clientNotifier := range u.clientNotifiers {
		clientNotifier.notifyTransaction(ct, transaction)
	}
}

func newUserNotifier(dataCollector obs.DataCollector, userID uint64) *UserNotifier {
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
