package realtime

import (
	"encoding/json"
	"log"

	"github.com/teamyapp/cloud/libs/connection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

// Task[Create, Update, Delete] -> :::: Team -> User -> Client
// Team[Update, Delete] -> :::: Team -> User -> Client
// Invitation[Create, Update, Delete] -> :::: Team -> User -> Client
// Message[Create, Update, Delete] -> Thread -> Task -> :::: Team -> User -> Client
// User[Update, Delete] -> TeamMember -> :::: Team -> User -> Client
// TeamMember[Create, Delete] -> :::: Team -> User -> Client
// TaskAwaitForRelation[Create, Delete] -> :::: Task -> Team -> User -> Client

// Team -> User -> Client

// Client connect
// 1) link client to user ID through access token
// 2) initialize all team level subscriptions for the user

// Client disconnect
// 1) Remove client for that user
// 2) Clean up user and team if no subscription

const stateSyncerBufferSize = 50

type Mutation struct {
	ID             uint64
	CollectionType CollectionType
	MutationType   MutationType
	TeamIDs        []uint64
	Payload        interface{}
}

type StateSyncer struct {
	teamMemberDao  dao.TeamMember
	mutations      chan Mutation
	teamNotifiers  map[uint64]*TeamNotifier
	userNotifiers  map[uint64]*UserNotifier
	nextClientID   uint64
	nextMutationID uint64
}

func (s *StateSyncer) OnClientConnect(userID uint64, conn connection.Connection) error {
	log.Printf("client connected: userID=%v, clientID=%v\n", userID, s.nextClientID)
	userNotifier, err := s.getUserNotifier(userID)
	if err != nil {
		return err
	}

	clientNotifier := newClientNotifier(conn, s.nextClientID)
	userNotifier.registerClientNotifier(s.nextClientID, clientNotifier)
	s.nextClientID++
	return nil
}

func (s *StateSyncer) NotifyMutation(mutation Mutation) {
	mutation.ID = s.nextMutationID
	s.nextMutationID++
	s.mutations <- mutation
}

func (s *StateSyncer) newUserNotifier(userID uint64) (*UserNotifier, error) {
	userNotifier := newUserNotifier(userID)
	go func() {
		<-userNotifier.subscribeUserDisconnect()
		delete(s.userNotifiers, userID)
	}()
	s.userNotifiers[userID] = userNotifier
	err := s.subscribeToTeams(userID, userNotifier)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return userNotifier, nil
}

func (s *StateSyncer) subscribeToTeams(userID uint64, userNotifier *UserNotifier) error {
	teamIDs, err := s.teamMemberDao.FindTeamIDsByUserID(userID)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, teamID := range teamIDs {
		teamNotifier, ok := s.teamNotifiers[teamID]
		if !ok {
			teamNotifier = s.newTeamNotifier(teamID)
		}

		log.Printf("subscribed to team: teamID=%v, userID=%v\n", teamID, userID)
		teamNotifier.registerUserNotifier(userID, userNotifier)
	}

	return nil
}

func (s *StateSyncer) newTeamNotifier(teamID uint64) *TeamNotifier {
	teamNotifier, ok := s.teamNotifiers[teamID]
	if !ok {
		teamNotifier = newTeamNotifier(teamID)
		go func() {
			<-teamNotifier.subscribeTeamDisconnect()
			delete(s.teamNotifiers, teamID)
		}()
		s.teamNotifiers[teamID] = teamNotifier
	}

	return teamNotifier
}

func (s StateSyncer) getUserNotifier(userID uint64) (*UserNotifier, error) {
	userNotifier, ok := s.userNotifiers[userID]
	var err error
	if !ok {
		userNotifier, err = s.newUserNotifier(userID)
		if err != nil {
			log.Println(err)
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
	teamMemberDao dao.TeamMember,
) *StateSyncer {
	stateSyncer := &StateSyncer{
		teamMemberDao:  teamMemberDao,
		mutations:      make(chan Mutation, stateSyncerBufferSize),
		teamNotifiers:  map[uint64]*TeamNotifier{},
		userNotifiers:  map[uint64]*UserNotifier{},
		nextClientID:   1,
		nextMutationID: 1,
	}
	go func() {
		for mutation := range stateSyncer.mutations {
			buf, err := json.MarshalIndent(mutation, "", "    ")
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("StateSyncer processing mutation: mutation=%v\n", string(buf))
			}

			for _, teamID := range mutation.TeamIDs {
				if mutation.CollectionType == TeamMemberCollectionType &&
					mutation.MutationType == CreateMutationType {
					teamMember := mutation.Payload.(entity.TeamMember)
					userNotifier, err := stateSyncer.getUserNotifier(teamMember.UserID)
					if err != nil {
						log.Println(err)
					} else {
						err = stateSyncer.subscribeToTeams(teamMember.UserID, userNotifier)
						if err != nil {
							log.Println(err)
						}
					}
				}

				stateSyncer.notifyTeam(teamID, mutation)
			}
		}
	}()
	return stateSyncer
}
