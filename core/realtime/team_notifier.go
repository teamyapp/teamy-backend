package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
)

type TeamNotifier struct {
	dataCollector             obs.DataCollector
	teamID                    uint64
	teamDisconnectCh          chan bool
	teamDisconnectSubscribers []chan bool
	userNotifiers             map[uint64]*UserNotifier
}

func (t TeamNotifier) registerUserNotifier(userID uint64, userNotifier *UserNotifier) {
	t.userNotifiers[userID] = userNotifier
}

func (t TeamNotifier) unregisterUserNotifier(userID uint64) {
	delete(t.userNotifiers, userID)
	if len(t.userNotifiers) == 0 {
		t.teamDisconnectCh <- true
	}
}

func (t *TeamNotifier) subscribeTeamDisconnect() chan bool {
	subscriber := make(chan bool)
	t.teamDisconnectSubscribers = append(t.teamDisconnectSubscribers, subscriber)
	return subscriber
}

func (t TeamNotifier) notifyTransaction(ct context.Context, transaction *Transaction) {
	t.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"summary": "notify transaction",
			"teamId":  t.teamID,
		},
	})
	for _, userNotifier := range t.userNotifiers {
		userNotifier.notifyTransaction(ct, transaction)
	}
}

func newTeamNotifier(dataCollector obs.DataCollector, teamID uint64) *TeamNotifier {
	teamNotifier := &TeamNotifier{
		dataCollector:             dataCollector,
		teamID:                    teamID,
		userNotifiers:             map[uint64]*UserNotifier{},
		teamDisconnectCh:          make(chan bool),
		teamDisconnectSubscribers: make([]chan bool, 0),
	}
	go func() {
		<-teamNotifier.teamDisconnectCh
		for _, subscriber := range teamNotifier.teamDisconnectSubscribers {
			subscriber <- true
		}
	}()
	return teamNotifier
}
