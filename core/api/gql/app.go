package gql

import (
	"github.com/graph-gophers/graphql-go"
)

type App struct {
	deps Dependencies
}

func (a App) ID() graphql.ID {
	panic("implement me")
}

func (a App) Name() string {
	panic("implement me")
}

func (a App) IconURL() string {
	panic("implement me")
}

func (a App) Creator() (User, error) {
	panic("implement me")
}

func (a App) IncludeUI() bool {
	panic("implement me")
}

func (a App) UIEntryPointURL() *string {
	panic("implement me")
}

func (a App) IsPublic() bool {
	panic("implement me")
}

func (a App) CreatedAt() graphql.Time {
	panic("implement me")
}

func (a App) UpdatedAt() *graphql.Time {
	panic("implement me")
}

func (a App) Installations() ([]AppInstallation, error) {
	panic("implement me")
}

func (a App) InstallationCount() (int32, error) {
	panic("implement me")
}

func newApp(deps Dependencies) App {
	return App{deps: deps}
}
