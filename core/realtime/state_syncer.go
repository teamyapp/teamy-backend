package realtime

import (
	"context"
	"errors"
	"sync"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

// Client connect
// 1) link client to user ID through access token
// 2) initialize all team level subscriptions for the user

// Client disconnect
// 1) Remove client for that user
// 2) Clean up user and team if no subscription

type StateSyncer struct {
	dataCollector           obs.DataCollector
	teamMemberDao           dao.TeamMember
	teamNotifiers           map[uint64]*TeamNotifier
	userNotifiers           map[uint64]*UserNotifier
	nextClientID            uint64
	nextMutationID          uint64
	nextTransactionID       uint64
	nextClientTransactionID uint64
	transactionMut          *sync.Mutex
}

func (s *StateSyncer) BeginTransaction() {
	s.transactionMut.Lock()
}

func (s *StateSyncer) EndTransaction() {
	s.transactionMut.Unlock()
}

func (s *StateSyncer) OnClientConnect(userID uint64, conn connection.Connection) error {
	ct := ctx.WithClientID(context.Background(), s.nextClientID)
	s.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"Summary": "client connected",
			"UserID":  userID,
		},
	})
	userNotifier, err := s.GetUserNotifier(ct, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	clientNotifier := newClientNotifier(s.dataCollector, conn, s.nextClientID)
	userNotifier.registerClientNotifier(s.nextClientID, clientNotifier)
	clientNotifier.sentMetadata()
	s.nextClientID++
	return nil
}

func (s *StateSyncer) NextMutationID() uint64 {
	mutationID := s.nextMutationID
	s.nextMutationID++
	return mutationID
}

func (s *StateSyncer) NextTransactionID() uint64 {
	transactionID := s.nextTransactionID
	s.nextTransactionID++
	return transactionID
}

func (s *StateSyncer) NextClientTransactionID() uint64 {
	clientTransactionID := s.nextClientTransactionID
	s.nextClientTransactionID++
	return clientTransactionID
}

func (s *StateSyncer) OnInitialStateReady(userID uint64, clientID uint64) error {
	ct := ctx.WithClientID(context.Background(), s.nextClientID)
	userNotifier, ok := s.userNotifiers[userID]
	if !ok {
		err := errors.New("userNotifier not found")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"UserID": userID,
			},
		})
		return err
	}

	clientNotifier, ok := userNotifier.clientNotifiers[clientID]
	if !ok {
		err := errors.New("clientNotifier not found")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"ClientID": clientID,
			},
		})
		return err
	}

	clientNotifier.onInitialStateReady()
	return nil
}

func (s *StateSyncer) newUserNotifier(ct context.Context, userID uint64) (*UserNotifier, error) {
	userNotifier := newUserNotifier(s.dataCollector, userID)
	go func() {
		<-userNotifier.subscribeUserDisconnect()
		delete(s.userNotifiers, userID)
	}()
	s.userNotifiers[userID] = userNotifier
	err := s.SubscribeToTeams(ct, userID, userNotifier)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return userNotifier, nil
}

func (s *StateSyncer) SubscribeToTeams(ct context.Context, userID uint64, userNotifier *UserNotifier) error {
	teamIDs, err := s.teamMemberDao.FindTeamIDsByUserID(ct, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	for _, teamID := range teamIDs {
		teamNotifier, ok := s.teamNotifiers[teamID]
		if !ok {
			teamNotifier = s.newTeamNotifier(teamID)
		}

		s.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
			obs.MessageProp: obs.Props{
				"Summary": "subscribed to team",
				"TeamID":  teamID,
				"UserID":  userID,
			},
		})
		teamNotifier.registerUserNotifier(userID, userNotifier)
	}

	return nil
}

func (s *StateSyncer) newTeamNotifier(teamID uint64) *TeamNotifier {
	teamNotifier, ok := s.teamNotifiers[teamID]
	if !ok {
		teamNotifier = newTeamNotifier(s.dataCollector, teamID)
		go func() {
			<-teamNotifier.subscribeTeamDisconnect()
			delete(s.teamNotifiers, teamID)
		}()
		s.teamNotifiers[teamID] = teamNotifier
	}

	return teamNotifier
}

func (s StateSyncer) GetUserNotifier(ct context.Context, userID uint64) (*UserNotifier, error) {
	userNotifier, ok := s.userNotifiers[userID]
	var err error
	if !ok {
		userNotifier, err = s.newUserNotifier(ct, userID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return nil, err
		}
	}

	return userNotifier, nil
}

func (s *StateSyncer) GetTeamNotifier(ct context.Context, teamID uint64) (*TeamNotifier, error) {
	teamNotifier, ok := s.teamNotifiers[teamID]
	if !ok {
		err := errors.New("teamNotifier not found")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"TeamID": teamID,
			},
		})
		return nil, err
	}

	return teamNotifier, nil
}

func (s *StateSyncer) GetClientNotifiersByUserID(ct context.Context, userID uint64) ([]*ClientNotifier, error) {
	teamIDs, err := s.teamMemberDao.FindTeamIDsByUserID(ct, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []*ClientNotifier{}, err
	}

	clientNotifiersMap := make(map[uint64]*ClientNotifier)
	for _, teamID := range teamIDs {
		teamClientNotifiers, err := s.GetClientNotifiersByTeamID(ct, teamID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return []*ClientNotifier{}, err
		}

		for _, clientNotifier := range teamClientNotifiers {
			clientNotifiersMap[clientNotifier.clientID] = clientNotifier
		}
	}

	clientNotifiers := make([]*ClientNotifier, 0)
	for _, clientNotifier := range clientNotifiersMap {
		clientNotifiers = append(clientNotifiers, clientNotifier)
	}

	return clientNotifiers, nil
}

func (s *StateSyncer) GetClientNotifiersByTeamID(ct context.Context, teamID uint64) ([]*ClientNotifier, error) {
	teamNotifier, err := s.GetTeamNotifier(ct, teamID)
	if err != nil {
		return nil, err
	}

	clientNotifiers := make([]*ClientNotifier, 0)
	for _, userNotifier := range teamNotifier.GetUserNotifiers() {
		for _, clientNotifier := range userNotifier.GetClientNotifiers() {
			clientNotifiers = append(clientNotifiers, clientNotifier)
		}
	}

	return clientNotifiers, nil
}

func NewStateSyncer(
	dataCollector obs.DataCollector,
	teamMemberDao dao.TeamMember,
) *StateSyncer {
	stateSyncer := &StateSyncer{
		dataCollector:           dataCollector,
		teamMemberDao:           teamMemberDao,
		teamNotifiers:           map[uint64]*TeamNotifier{},
		userNotifiers:           map[uint64]*UserNotifier{},
		nextClientID:            1,
		nextMutationID:          1,
		nextTransactionID:       1,
		nextClientTransactionID: 1,
		transactionMut:          new(sync.Mutex),
	}

	return stateSyncer
}
