package realtime

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

// Client connect
// 1) link client to user ID through access token
// 2) initialize all team level subscriptions for the user

// Client disconnect
// 1) Remove client for that user
// 2) Clean up user and team if no subscription

const stateSyncerBufferSize = 50

type StateSyncer struct {
	dataCollector  obs.DataCollector
	teamMemberDao  dao.TeamMember
	mutations      chan Mutation
	teamNotifiers  map[uint64]*TeamNotifier
	userNotifiers  map[uint64]*UserNotifier
	nextClientID   uint64
	nextMutationID uint64
}

func (s *StateSyncer) OnClientConnect(userID uint64, conn connection.Connection) error {
	ct := ctx.WithClientID(context.Background(), s.nextClientID)
	s.dataCollector.Logger.LogWithContext(ct, obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"summary": "client connected",
			"userID":  userID,
		},
	})
	userNotifier, err := s.getUserNotifier(ct, userID)
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

func (s *StateSyncer) NotifyMutation(mutation Mutation) {
	mutation.ID = s.nextMutationID
	s.nextMutationID++
	s.mutations <- mutation
}

func (s *StateSyncer) OnInitialStateReady(userID uint64, clientID uint64) error {
	ct := ctx.WithClientID(context.Background(), s.nextClientID)
	userNotifier, ok := s.userNotifiers[userID]
	if !ok {
		err := errors.New("userNotifier not found")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"userID": userID,
			},
		})
		return err
	}

	clientNotifier, ok := userNotifier.clientNotifiers[clientID]
	if !ok {
		err := errors.New("clientNotifier not found")
		s.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"clientID": clientID,
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
	err := s.subscribeToTeams(ct, userID, userNotifier)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return userNotifier, nil
}

func (s *StateSyncer) subscribeToTeams(ct context.Context, userID uint64, userNotifier *UserNotifier) error {
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
				"summary": "subscribed to team",
				"teamID":  teamID,
				"userID":  userID,
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

func (s StateSyncer) getUserNotifier(ct context.Context, userID uint64) (*UserNotifier, error) {
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

func (s StateSyncer) notifyTeam(teamID uint64, mutation Mutation) {
	teamNotifier, ok := s.teamNotifiers[teamID]
	if !ok {
		return
	}

	teamNotifier.processMutation(mutation)
}

func NewStateSyncer(
	dataCollector obs.DataCollector,
	teamMemberDao dao.TeamMember,
) *StateSyncer {
	stateSyncer := &StateSyncer{
		dataCollector:  dataCollector,
		teamMemberDao:  teamMemberDao,
		mutations:      make(chan Mutation, stateSyncerBufferSize),
		teamNotifiers:  map[uint64]*TeamNotifier{},
		userNotifiers:  map[uint64]*UserNotifier{},
		nextClientID:   1,
		nextMutationID: 1,
	}
	go func() {
		ct := context.Background()
		for mutation := range stateSyncer.mutations {
			newCtx := WithMutationID(ct, mutation.ID)
			buf, err := json.Marshal(mutation)
			if err != nil {
				dataCollector.Logger.LogWithContext(newCtx, obs.Error, obs.Props{obs.CauseProp: err})
			} else {
				dataCollector.Logger.LogWithContext(newCtx, obs.Info, obs.Props{
					obs.MessageProp: obs.Props{
						"summary":  "process mutation",
						"mutation": string(buf),
					},
				})
			}

			for _, teamID := range mutation.TeamIDs {
				if mutation.CollectionType == TeamMemberCollectionType &&
					mutation.MutationType == CreateMutationType {
					teamMember := mutation.Payload.(entity.TeamMember)
					newCtx := WithMutationID(ct, mutation.ID)
					userNotifier, err := stateSyncer.getUserNotifier(newCtx, teamMember.UserID)
					if err != nil {
						dataCollector.Logger.LogWithContext(newCtx, obs.Error, obs.Props{obs.CauseProp: err})
					} else {
						err = stateSyncer.subscribeToTeams(newCtx, teamMember.UserID, userNotifier)
						if err != nil {
							dataCollector.Logger.LogWithContext(newCtx, obs.Error, obs.Props{obs.CauseProp: err})
						}
					}
				}

				stateSyncer.notifyTeam(teamID, mutation)
			}
		}
	}()
	return stateSyncer
}
