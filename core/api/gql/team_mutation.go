package gql

import (
	"context"
	"log"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/teamy-backend/core/entity"
)

func (m Mutation) CreateTeam(ct context.Context, args struct {
	Team struct {
		Name string
	}
}) (Team, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Team{}, err
	}

	genTeamIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamID"}
	genTeamIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamIDReq)
	if err != nil {
		log.Println(err)
		return Team{}, err
	}

	team := entity.Team{
		ID:            genTeamIDRes.UniqueNumber,
		Name:          args.Team.Name,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now(),
	}
	err = m.deps.teamSyncer.CreateAndSyncTeam(team)
	if err != nil {
		return Team{}, err
	}

	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(team.ID, userID)
	if err != nil {
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
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return Team{}, err
	}

	team, err := m.deps.teamDao.FindTeamByID(teamID)
	if err != nil {
		return Team{}, err
	}

	team.Name = args.Input.Name
	ownerUserID, err := fromGraphQLID(args.Input.OwnerUserID)
	if err != nil {
		return Team{}, err
	}

	team.OwnerUserID = ownerUserID
	updatedAt := time.Now()
	team.UpdatedAt = &updatedAt
	err = m.deps.teamSyncer.UpdateAndSyncTeam(team)
	if err != nil {
		return Team{}, err
	}

	return newTeam(m.deps, team), err
}

func (m Mutation) CreateTeamIconUploadSession(ct context.Context, args struct {
	TeamID graphql.ID
}) (graphql.ID, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		log.Println(err)
		return "", err
	}

	uploadSessionID, err := m.deps.teamService.CreateTeamIconUploadSession(ct, teamID)
	if err != nil {
	    log.Println(err)
		return "", err
	}

	return toGraphQLID(uploadSessionID), nil
}

func (m Mutation) FinishTeamIconUploadSession(ct context.Context, args struct {
	TeamID              graphql.ID
	FileUploadSessionID graphql.ID
}) (graphql.ID, error) {
	fileUploadSessionID, err := fromGraphQLID(args.FileUploadSessionID)
	if err != nil {
		log.Println(err)
		return "", err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		log.Println(err)
		return "", err
	}

	uploadSessionID, err := m.deps.teamService.FinishTeamIconUploadSession(ct, teamID, fileUploadSessionID)
	if err != nil {
        log.Println(err)
		return "", err
	}

	return toGraphQLID(uploadSessionID), nil
}

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (User, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(teamID, memberUserID)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(memberUserID)
	if err != nil {
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
		return User{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		return User{}, err
	}

	// TODO: ensure user is inside the team
	err = m.deps.teamMemberSyncer.DeleteAndSyncTeamMember(teamID, memberUserID)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(memberUserID)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}
