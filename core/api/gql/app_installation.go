package gql

import (
	"github.com/graph-gophers/graphql-go"
)

type AppInstallation struct {
	deps Dependencies
}

func (a AppInstallation) ID() graphql.ID {
	panic("implement me")
}

func (a AppInstallation) App() (App, error) {
	panic("implement me")
}

func (a AppInstallation) Secret() string {
	panic("implement me")
}

func (a AppInstallation) InstalledTeam() (Team, error) {
	panic("implement me")
}

func (a AppInstallation) InstalledBy() (User, error) {
	panic("implement me")
}

func (a AppInstallation) InstalledAt() graphql.Time {
	panic("implement me")
}

func newAppInstallation(deps Dependencies) AppInstallation {
	return AppInstallation{deps: deps}
}
