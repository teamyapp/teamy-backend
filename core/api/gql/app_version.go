package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type AppVersion struct {
}

func (a AppVersion) Number(ctx context.Context) int32 {
	panic("not implemented")
}

func (a AppVersion) AppName(ctx context.Context) string {
	panic("not implemented")
}

func (a AppVersion) Description(ctx context.Context) string {
	panic("not implemented")
}

func (a AppVersion) Changes(ctx context.Context) []string {
	panic("not implemented")
}

func (a AppVersion) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (a AppVersion) CreatedBy(ctx context.Context) User {
	panic("not implemented")
}

func (a AppVersion) Prices(ctx context.Context) []Money {
	panic("not implemented")
}

func (a AppVersion) IsReady(ctx context.Context) bool {
	panic("not implemented")
}

func (a AppVersion) App(ctx context.Context) App {
	panic("not implemented")
}

func (m Mutation) CreateAppVersion(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	}) AppVersion {
	panic("not implemented")
}

func (m Mutation) CreateAppPackageUploadSession(
	ctx context.Context,
	args struct {
		AppID         graphql.ID
		VersionNumber int32
	}) graphql.ID {
	panic("not implemented")
}

func (m Mutation) FinishAppPackageUploadSession(
	ctx context.Context,
	args struct {
		AppID               graphql.ID
		VersionNumber       int32
		FileUploadSessionID graphql.ID
	}) AppVersion {
	panic("not implemented")
}

func (m Mutation) DeleteAppVersion(
	ctx context.Context,
	args struct {
		AppID         graphql.ID
		VersionNumber int32
	}) AppVersion {
	panic("not implemented")
}
