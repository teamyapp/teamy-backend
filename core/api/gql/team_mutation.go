package gql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
)

func (m Mutation) CreateTeam(ct context.Context, args struct {
	Team struct {
		Name string
	}
}) (Team, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	genTeamIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamID"}
	genTeamIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamIDReq)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	team := entity.Team{
		ID:            genTeamIDRes.UniqueNumber,
		Name:          args.Team.Name,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now(),
	}
	err = m.deps.teamSyncer.CreateAndSyncTeam(ct, team)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	teamMember := entity.TeamMember{
		TeamID:    team.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(ct, teamMember)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	return newTeam(m.deps, team), nil
}

func (m Mutation) UpdateTeam(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		Name        string
		OwnerUserID graphql.ID
	}
}) (Team, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewUpdateTeamSettingsQuery(userID, teamID)
		hasPermission, err := m.hasPermission(ct, query)
		if err != nil {
			return Team{}, err
		}

		if !hasPermission {
			return Team{}, ResolverError{
				Code:    unauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorize: %v", query),
			}
		}
	}

	team, err := m.deps.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	team.Name = args.Input.Name
	ownerUserID, err := fromGraphQLID(args.Input.OwnerUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	team.OwnerUserID = ownerUserID
	updatedAt := time.Now()
	team.UpdatedAt = &updatedAt
	err = m.deps.teamSyncer.UpdateAndSyncTeam(ct, team)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	return newTeam(m.deps, team), nil
}

func (m Mutation) CreateTeamIconUploadSession(ct context.Context, args struct {
	TeamID graphql.ID
}) (graphql.ID, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	uploadSessionID, err := m.deps.teamService.CreateTeamIconUploadSession(ct, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	return toGraphQLID(uploadSessionID), nil
}

func (m Mutation) FinishTeamIconUploadSession(ct context.Context, args struct {
	TeamID              graphql.ID
	FileUploadSessionID graphql.ID
}) (Team, error) {
	fileUploadSessionID, err := fromGraphQLID(args.FileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	team, err := m.deps.teamService.FinishTeamIconUploadSession(ct, teamID, fileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	return newTeam(m.deps, team), nil
}

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (User, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	teamMember := entity.TeamMember{
		TeamID:    teamID,
		UserID:    memberUserID,
		CreatedAt: time.Now(),
	}
	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(ct, teamMember)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(ct, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) RemoveMemberFromTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (User, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	// TODO: ensure user is inside the team
	err = m.deps.teamMemberSyncer.DeleteAndSyncTeamMember(ct, teamID, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(ct, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	return newUser(m.deps, user), nil
}
