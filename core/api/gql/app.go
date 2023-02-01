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

func (a App) APISecret() string {
	panic("implement me")
}

func (a App) ActiveVersion() (*AppVersion, error) {
	panic("implement me")
}

func (a App) AppName() string {
	panic("implement me")
}

func (a App) Versions() ([]AppVersion, error) {
	panic("implement me")
}

func (a App) TeamInstallations() ([]AppTeamInstallation, error) {
	panic("implement me")
}

func (a App) InstallationCount() int32 {
	panic("implement me")
}

func (a App) Description() string {
	panic("implement me")
}

func (a App) Creator() (User, error) {
	panic("implement me")
}

func (a App) CreatedAt() graphql.Time {
	panic("implement me")
}

func (a App) UpdatedAt() *graphql.Time {
	panic("implement me")
}

func newApp(deps Dependencies) App {
	return App{deps: deps}
}
