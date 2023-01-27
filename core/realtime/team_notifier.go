package realtime

import (
	"github.com/teamyapp/cloud/libs/telemetry"
)

type TeamNotifier struct {
	dataCollector             telemetry.DataCollector
	teamID                    uint64
	teamDisconnectCh          chan bool
	teamDisconnectSubscribers []chan bool
	userNotifiers             map[uint64]*UserNotifier
}

func (t TeamNotifier) GetUserNotifiers() map[uint64]*UserNotifier {
	return t.userNotifiers
}

func (t TeamNotifier) registerUserNotifier(userID uint64, userNotifier *UserNotifier) {
	t.userNotifiers[userID] = userNotifier
}

func (t TeamNotifier) UnregisterUserNotifier(userID uint64) {
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

func newTeamNotifier(dataCollector telemetry.DataCollector, teamID uint64) *TeamNotifier {
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
