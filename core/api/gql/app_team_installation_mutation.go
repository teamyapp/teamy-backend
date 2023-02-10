package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateAppTeamInstallation(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        graphql.ID
}) (AppTeamInstallation, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppTeamInstallation{}, errs.ToResolverErr(internalErr)
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppTeamInstallation{}, errs.ToResolverErr(internalErr)
	}

	appTeamInstallation, err := m.deps.appService.CreateAppInstallation(ct, teamID, appID, args.VersionNumber)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppTeamInstallation{}, errs.ToResolverErr(err)
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
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppTeamInstallation{}, errs.ToResolverErr(internalErr)
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppTeamInstallation{}, errs.ToResolverErr(internalErr)
	}

	input := service.UpdateAppTeamInstallationInput{
		EnabledVersionNumber: args.Input.EnabledVersionNumber,
	}
	appTeamInstallation, err := m.deps.appService.UpdateAppInstallation(ct, appID, teamID, input)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppTeamInstallation{}, errs.ToResolverErr(err)
	}

	return newAppTeamInstallation(m.deps, appTeamInstallation), nil
}

func (m Mutation) DeleteAppTeamInstallation(ct context.Context, args struct {
	AppID  graphql.ID
	TeamID graphql.ID
}) (AppTeamInstallation, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppTeamInstallation{}, errs.ToResolverErr(internalErr)
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppTeamInstallation{}, errs.ToResolverErr(internalErr)
	}

	appTeamInstallation, err := m.deps.appService.DeleteAppInstallation(ct, appID, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppTeamInstallation{}, errs.ToResolverErr(err)
	}

	return newAppTeamInstallation(m.deps, appTeamInstallation), nil
}
