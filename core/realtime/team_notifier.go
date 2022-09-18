package realtime

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
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

func (t TeamNotifier) processMutation(mutation Mutation) {

	t.dataCollector.Logger.Log(obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"summary":    "client disconnected",
			"teamID":     t.teamID,
			"mutationID": mutation.ID,
		},
	})
	for _, userNotifier := range t.userNotifiers {
		userNotifier.processMutation(mutation)
	}

	if mutation.CollectionType == TeamMemberCollectionType &&
		mutation.MutationType == DeleteMutationType {
		teamMember := mutation.Payload.(entity.TeamMember)
		t.unregisterUserNotifier(teamMember.UserID)
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
