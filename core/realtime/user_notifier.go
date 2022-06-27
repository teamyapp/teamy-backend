package realtime

import (
	"log"
)

type UserNotifier struct {
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
		log.Printf("client disconnected: clientID=%v\n", clientID)
		u.unregisterClientNotifier(clientID)
	}()
}

func (u UserNotifier) unregisterClientNotifier(clientID uint64) {
	delete(u.clientNotifiers, clientID)
	if len(u.clientNotifiers) == 0 {
		u.userDisconnectCh <- true
	}
}

func (u UserNotifier) processMutation(mutation Mutation) {
	log.Printf("UserNotifier processing mutation: userID=%v mutationID=%v\n", u.userID, mutation.ID)
	for _, clientNotifier := range u.clientNotifiers {
		clientNotifier.processMutation(mutation)
	}
}

func newUserNotifier(userID uint64) *UserNotifier {
	userNotifier := &UserNotifier{
		userID:                    userID,
		clientNotifiers:           map[uint64]*ClientNotifier{},
		userDisconnectSubscribers: make([]chan bool, 0),
	}
	go func() {
		<-userNotifier.userDisconnectCh
		for _, subscriber := range userNotifier.userDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return userNotifier
}
