package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserGroup interface {
	Group
	Rollouts(ctx context.Context) []UserRollout
}

type StaticUserGroup struct {
}

var _ UserGroup = (*StaticUserGroup)(nil)

func (s StaticUserGroup) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (s StaticUserGroup) Type(ctx context.Context) entity.GroupType {
	panic("not implemented")
}

func (s StaticUserGroup) Name(ctx context.Context) string {
	panic("not implemented")
}

func (s StaticUserGroup) Users(ctx context.Context) []User {
	panic("not implemented")
}

func (s StaticUserGroup) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (s StaticUserGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (s StaticUserGroup) Rollouts(ctx context.Context) []UserRollout {
	panic("not implemented")
}

func (s StaticUserGroup) App(ctx context.Context) App {
	panic("not implemented")
}

type FilterUserGroup struct {
}

var _ UserGroup = (*FilterUserGroup)(nil)
var _ FilterGroup = (*FilterUserGroup)(nil)

func (f FilterUserGroup) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (f FilterUserGroup) Type(ctx context.Context) entity.GroupType {
	panic("not implemented")
}

func (f FilterUserGroup) Name(ctx context.Context) string {
	panic("not implemented")
}

func (f FilterUserGroup) Filter(ctx context.Context) string {
	panic("not implemented")
}

func (f FilterUserGroup) UserCount(ctx context.Context) int32 {
	panic("not implemented")
}

func (f FilterUserGroup) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (f FilterUserGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (f FilterUserGroup) Rollouts(ctx context.Context) []UserRollout {
	panic("not implemented")
}

func (f FilterUserGroup) App(ctx context.Context) App {
	panic("not implemented")
}

func (m Mutation) CreateStaticUserGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name    string
			UserIDs []graphql.ID
		}
	},
) StaticUserGroup {
	panic("not implemented")
}

func (m Mutation) UpdateStaticUserGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name    string
			UserIDs []graphql.ID
		}
	},
) StaticUserGroup {
	panic("not implemented")
}

func (m Mutation) CreateFilterUserGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name   string
			Filter string
		}
	},
) FilterUserGroup {
	panic("not implemented")
}

func (m Mutation) UpdateFilterUserGroup(
	ctx context.Context,
	args struct {
		GroupID graphql.ID
		Input   struct {
			Name   string
			Filter string
		}
	},
) FilterUserGroup {
	panic("not implemented")
}
