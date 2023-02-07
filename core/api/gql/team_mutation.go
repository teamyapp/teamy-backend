package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateTeam(ct context.Context, args struct {
	Team struct {
		Name string
	}
}) (Team, error) {
	createTeamInput := service.CreateTeamInput{
		Name: args.Team.Name,
	}
	team, err := m.deps.teamService.CreateTeam(ct, createTeamInput)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return Team{}, errs.ToResolverErr(err)
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
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return Team{}, errs.ToResolverErr(internalErr)
	}

	ownerUserID, argErr := fromGraphQLID(args.Input.OwnerUserID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return Team{}, errs.ToResolverErr(internalErr)
	}

	updateTeamInput := service.UpdateTeamInput{
		Name:        args.Input.Name,
		OwnerUserID: ownerUserID,
	}
	team, err := m.deps.teamService.UpdateTeam(ct, teamID, updateTeamInput)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(m.deps, team), nil
}

func (m Mutation) CreateTeamIconUploadSession(ct context.Context, args struct {
	TeamID graphql.ID
}) (graphql.ID, error) {
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return "", errs.ToResolverErr(internalErr)
	}

	uploadSessionID, err := m.deps.teamService.CreateTeamIconUploadSession(ct, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", errs.ToResolverErr(err)
	}

	return toGraphQLID(uploadSessionID), nil
}

func (m Mutation) FinishTeamIconUploadSession(ct context.Context, args struct {
	TeamID              graphql.ID
	FileUploadSessionID graphql.ID
}) (Team, error) {
	fileUploadSessionID, argErr := fromGraphQLID(args.FileUploadSessionID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return Team{}, errs.ToResolverErr(internalErr)
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return Team{}, errs.ToResolverErr(internalErr)
	}

	team, err := m.deps.teamService.FinishTeamIconUploadSession(ct, teamID, fileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(m.deps, team), nil
}
