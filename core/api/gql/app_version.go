package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type AppVersion struct {
	deps *Dependencies
}

func (a AppVersion) App(ct context.Context) (App, error) {
	panic("implement me")
}

func (a AppVersion) VersionNumber() int32 {
	panic("implement me")
}

func (a AppVersion) Name() string {
	panic("implement me")
}

func (a AppVersion) IconURL() *string {
	panic("implement me")
}

func (a AppVersion) HasUiExtension() bool {
	panic("implement me")
}

func (a AppVersion) UiExtensionEntryPointPath() *string {
	panic("implement me")
}

func (a AppVersion) IsPublic() bool {
	panic("implement me")
}

func (a AppVersion) VisibleToTeams() []Team {
	panic("implement me")
}

func (a AppVersion) CreatedAt() graphql.Time {
	panic("implement me")
}

func (a AppVersion) UpdatedAt() *graphql.Time {
	panic("implement me")
}

func newAppVersion(deps *Dependencies) AppVersion {
	return AppVersion{deps: deps}
}
