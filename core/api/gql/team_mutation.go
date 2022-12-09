package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/obs"
)

func (m Mutation) CreateTeam(ct context.Context, args struct {
	Team struct {
		Name string
	}
}) (Team, error) {
	team, err := m.deps.teamService.CreateTeam(ct, args.Team.Name)
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
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	ownerUserID, err := fromGraphQLID(args.Input.OwnerUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Team{}, err
	}

	team, err := m.deps.teamService.UpdateTeam(ct, teamID, args.Input.Name, ownerUserID)
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
