package gql

import (
	"github.com/graph-gophers/graphql-go"
)

type AppTeamInstallation struct {
	deps Dependencies
}

func (a AppTeamInstallation) App() (App, error) {
	panic("implement me")
}

func (a AppTeamInstallation) EnabledVersion() (AppVersion, error) {
	panic("implement me")
}

func (a AppTeamInstallation) InstalledTeam() (Team, error) {
	panic("implement me")
}

func (a AppTeamInstallation) InstalledBy() (*User, error) {
	panic("implement me")
}

func (a AppTeamInstallation) InstalledAt() graphql.Time {
	panic("implement me")
}

func newAppTeamInstallation(deps Dependencies) AppTeamInstallation {
	return AppTeamInstallation{deps: deps}
}
