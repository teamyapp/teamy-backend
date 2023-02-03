package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateAppTeamInstallation(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        graphql.ID
}) (AppTeamInstallation, error) {
	appID, err := fromGraphQLID(args.AppID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	appTeamInstallation, err := m.deps.appService.CreateAppInstallation(ct, teamID, appID, args.VersionNumber)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	return newAppTeamInstallation(m.deps, appTeamInstallation), nil
}

func (m Mutation) UpdateAppTeamInstallation(ct context.Context, args struct {
	AppID  graphql.ID
	TeamID graphql.ID
	Input  struct {
		EnabledVersionNumber int32
	}
}) (AppTeamInstallation, error) {
	appID, err := fromGraphQLID(args.AppID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	input := service.UpdateAppTeamInstallationInput{
	    EnabledVersionNumber: args.Input.EnabledVersionNumber,
	}
	appTeamInstallation, err := m.deps.appService.UpdateAppInstallation(ct, appID, teamID, input)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	return newAppTeamInstallation(m.deps, appTeamInstallation), nil
}

func (m Mutation) DeleteAppTeamInstallation(ct context.Context, args struct {
	AppID  graphql.ID
	TeamID graphql.ID
}) (AppTeamInstallation, error) {
	appID, err := fromGraphQLID(args.AppID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	appTeamInstallation, err := m.deps.appService.DeleteAppInstallation(ct, appID, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppTeamInstallation{}, err
	}

	return newAppTeamInstallation(m.deps, appTeamInstallation), nil
}
