package realtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
)

// Client connect
// 1) link client to user ID through access token
// 2) initialize all team level subscriptions for the user

// Client disconnect
// 1) Remove client for that user
// 2) Clean up user and team if no subscription

type StateSyncer struct {
	dataCollector           telemetry.DataCollector
	teamMemberDao           dao.TeamMember
	teamNotifiers           map[uint64]*TeamNotifier
	userNotifiers           map[uint64]*UserNotifier
	nextClientID            uint64
	nextMutationID          uint64
	nextTransactionID       uint64
	nextClientTransactionID uint64
	transactionMut          *sync.Mutex
	clientIDMut             *sync.Mutex
	mutationIDMut           *sync.Mutex
	transactionIDMut        *sync.Mutex
	clientTransactionIDMut  *sync.Mutex
}

func (s *StateSyncer) BeginTransaction() {
	s.transactionMut.Lock()
}

func (s *StateSyncer) EndTransaction() {
	s.transactionMut.Unlock()
}

func (s *StateSyncer) OnClientConnect(userID uint64, conn connection.Connection) *errs.Error {
	ct := ctx.WithClientID(context.Background(), s.nextClientID)
	s.dataCollector.Logger.LogWithContext(ct, telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("client connected: userID=%v", userID),
	})
	userNotifier, err := s.GetUserNotifier(ct, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	s.clientIDMut.Lock()
	nextClientID := s.nextClientID
	s.nextClientID++
	s.clientIDMut.Unlock()

	clientNotifier := newClientNotifier(s.dataCollector, conn, nextClientID)
	userNotifier.registerClientNotifier(nextClientID, clientNotifier)
	clientNotifier.sentMetadata()

	return nil
}

func (s *StateSyncer) NextMutationID() uint64 {
	s.mutationIDMut.Lock()
	mutationID := s.nextMutationID
	s.nextMutationID++
	s.mutationIDMut.Unlock()
	return mutationID
}

func (s *StateSyncer) NextTransactionID() uint64 {
	s.transactionIDMut.Lock()
	transactionID := s.nextTransactionID
	s.nextTransactionID++
	s.transactionIDMut.Unlock()
	return transactionID
}

func (s *StateSyncer) NextClientTransactionID() uint64 {
	s.clientTransactionIDMut.Lock()
	clientTransactionID := s.nextClientTransactionID
	s.nextClientTransactionID++
	s.clientTransactionIDMut.Unlock()
	return clientTransactionID
}

func (s *StateSyncer) OnInitialStateReady(userID uint64, clientID uint64) *errs.Error {
	ct := ctx.WithClientID(context.Background(), s.nextClientID)
	userNotifier, ok := s.userNotifiers[userID]
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("userNotifier not found: userID=%v", userID),
		}
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	clientNotifier, ok := userNotifier.clientNotifiers[clientID]
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("clientNotifier not found: clientID=%v", clientID),
		}
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	clientNotifier.onInitialStateReady()
	return nil
}

func (s *StateSyncer) newUserNotifier(ct context.Context, userID uint64) (*UserNotifier, *errs.Error) {
	userNotifier := newUserNotifier(s.dataCollector, userID)
	go func() {
		<-userNotifier.subscribeUserDisconnect()
		delete(s.userNotifiers, userID)
	}()
	s.userNotifiers[userID] = userNotifier
	err := s.SubscribeToTeams(ct, userID, userNotifier)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return userNotifier, nil
}

func (s *StateSyncer) SubscribeToTeams(ct context.Context, userID uint64, userNotifier *UserNotifier) *errs.Error {
	teamIDs, err := s.teamMemberDao.FindTeamIDsByUserID(ct, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	for _, teamID := range teamIDs {
		teamNotifier, ok := s.teamNotifiers[teamID]
		if !ok {
			teamNotifier = s.newTeamNotifier(teamID)
		}

		s.dataCollector.Logger.LogWithContext(ct, telemetry.Info, telemetry.Props{
			telemetry.MessageProp: fmt.Sprintf("subscribed to team: teamID=%v, userID=%v",
				teamID,
				userID),
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

func (s StateSyncer) GetUserNotifier(ct context.Context, userID uint64) (*UserNotifier, *errs.Error) {
	userNotifier, ok := s.userNotifiers[userID]
	var err *errs.Error
	if !ok {
		userNotifier, err = s.newUserNotifier(ct, userID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return nil, err
		}
	}

	return userNotifier, nil
}

func (s *StateSyncer) GetTeamNotifier(ct context.Context, teamID uint64) (*TeamNotifier, *errs.Error) {
	teamNotifier, ok := s.teamNotifiers[teamID]
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("teamNotifier not found: teamID=%v", teamID),
		}
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return teamNotifier, nil
}

func (s *StateSyncer) GetAllClientNotifiersByUserID(ct context.Context, userID uint64) ([]*ClientNotifier, *errs.Error) {
	teamIDs, err := s.teamMemberDao.FindTeamIDsByUserID(ct, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return []*ClientNotifier{}, err
	}

	clientNotifiersMap := make(map[uint64]*ClientNotifier)
	for _, teamID := range teamIDs {
		teamClientNotifiers, err := s.GetClientNotifiersByTeamID(ct, teamID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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

func (s *StateSyncer) GetClientNotifiersByTeamID(ct context.Context, teamID uint64) ([]*ClientNotifier, *errs.Error) {
	teamNotifier, err := s.GetTeamNotifier(ct, teamID)
	if err != nil && err.Code == errs.NotFound {
		return nil, err
	}

	// There can be only 1 user client for a given team on a given device.
	clientNotifiers := make([]*ClientNotifier, 0)
	for _, userNotifier := range teamNotifier.GetUserNotifiers() {
		for _, clientNotifier := range userNotifier.GetClientNotifiers() {
			clientNotifiers = append(clientNotifiers, clientNotifier)
		}
	}

	return clientNotifiers, nil
}

func NewStateSyncer(
	dataCollector telemetry.DataCollector,
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
		clientIDMut:             new(sync.Mutex),
		mutationIDMut:           new(sync.Mutex),
		transactionIDMut:        new(sync.Mutex),
		clientTransactionIDMut:  new(sync.Mutex),
		transactionMut:          new(sync.Mutex),
	}

	return stateSyncer
}
