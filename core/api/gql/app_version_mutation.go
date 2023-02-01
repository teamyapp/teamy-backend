package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

func (m Mutation) CreateAppVersion(ct context.Context, args struct {
	AppID graphql.ID
}) (AppVersion, error) {
	panic("implement me")
}

func (m Mutation) UpdateAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	Input         struct {
		IconUrl                   *string
		HasUIExtension            bool
		UIExtensionEntrypointPath *string
		Changes                   *string
		IsPublic                  bool
	}
}) (AppVersion, error) {
	panic("implement me")
}

func (m Mutation) DeleteAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
}) (AppVersion, error) {
	panic("implement me")
}

func (m Mutation) AddVisibleTeamToAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        uint64
}) (AppVersion, error) {
	panic("implement me")
}

func (m Mutation) RemoveVisibleTeamFromAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        uint64
}) (AppVersion, error) {
	panic("implement me")
}
