package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

func (m Mutation) CreateAppTeamInstallation(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        graphql.ID
}) (AppTeamInstallation, error) {
	panic("implement me")
}

func (m Mutation) UpdateAppTeamInstallation(ct context.Context, args struct {
	AppID graphql.ID
	Input struct {
		EnabledVersionNumber *int32
	}
}) (AppTeamInstallation, error) {
	panic("implement me")
}

func (m Mutation) DeleteAppTeamInstallation(ct context.Context, args struct {
	AppID  graphql.ID
	TeamID graphql.ID
}) (AppTeamInstallation, error) {
	panic("implement me")
}
