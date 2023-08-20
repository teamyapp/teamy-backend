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

func newAppVersion(deps *Dependencies, appVersion entity) AppVersion {
	return AppVersion{}
}
