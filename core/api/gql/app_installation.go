package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type TeamAppInstallation struct {
}

func (t TeamAppInstallation) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (t TeamAppInstallation) InstalledTeam(ctx context.Context) Team {
	panic("not implemented")
}

func (t TeamAppInstallation) ActiveAppVersion(ctx context.Context) *AppVersion {
	panic("not implemented")
}
