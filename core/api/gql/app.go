package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type App struct {
}

func (a App) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (a App) Secrets(ctx context.Context) []AppSecret {
	panic("not implemented")
}

func (a App) TotalInstallations(ctx context.Context) int32 {
	panic("not implemented")
}

func (a App) Installations(ctx context.Context) []TeamAppInstallation {
	panic("not implemented")
}

func (a App) Versions(ctx context.Context) []AppVersion {
	panic("not implemented")
}

func (a App) UserGroups(ctx context.Context) []UserGroup {
	panic("not implemented")
}

func (a App) TeamGroups(ctx context.Context) []TeamGroup {
	panic("not implemented")
}

func (a App) UserRollouts(ctx context.Context) []UserRollout {
	panic("not implemented")
}

func (a App) TeamRollouts(ctx context.Context) []TeamRollout {
	panic("not implemented")
}

func (a App) OwnedByTeam(ctx context.Context) Team {
	panic("not implemented")
}

func (m Mutation) CreateApp(
	ctx context.Context,
	args struct {
		TeamID graphql.ID
		Name   string
	}) App {
	panic("not implemented")
}

func (m Mutation) DeleteApp(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	}) App {
	panic("not implemented")
}

func (m Mutation) InstallAppToTeam(
	ctx context.Context,
	args struct {
		AppID  graphql.ID
		TeamID graphql.ID
	}) TeamAppInstallation {
	panic("not implemented")
}

func (m Mutation) UninstallAppFromTeam(
	ctx context.Context,
	args struct {
		InstallationID graphql.ID
	}) TeamAppInstallation {
	panic("not implemented")
}
