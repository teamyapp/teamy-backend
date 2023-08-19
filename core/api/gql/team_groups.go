package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroup interface {
	Group
	Rollouts(ctx context.Context) []TeamRollout
}

type StaticTeamGroup struct {
}

func (s StaticTeamGroup) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (s StaticTeamGroup) Type(ctx context.Context) entity.GroupType {
	panic("not implemented")
}

func (s StaticTeamGroup) Name(ctx context.Context) string {
	panic("not implemented")
}

func (s StaticTeamGroup) Teams(ctx context.Context) []Team {
	panic("not implemented")
}

func (s StaticTeamGroup) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (s StaticTeamGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (s StaticTeamGroup) Rollouts(ctx context.Context) []TeamRollout {
	panic("not implemented")
}

func (s StaticTeamGroup) App(ctx context.Context) App {
	panic("not implemented")
}

type FilterTeamGroup struct {
}

var _ TeamGroup = (*FilterTeamGroup)(nil)
var _ FilterGroup = (*FilterTeamGroup)(nil)

func (f FilterTeamGroup) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (f FilterTeamGroup) Type(ctx context.Context) entity.GroupType {
	panic("not implemented")
}

func (f FilterTeamGroup) Name(ctx context.Context) string {
	panic("not implemented")
}

func (f FilterTeamGroup) Filter(ctx context.Context) string {
	panic("not implemented")
}

func (f FilterTeamGroup) TeamCount(ctx context.Context) int32 {
	panic("not implemented")
}

func (f FilterTeamGroup) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (f FilterTeamGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (f FilterTeamGroup) Rollouts(ctx context.Context) []TeamRollout {
	panic("not implemented")
}

func (f FilterTeamGroup) App(ctx context.Context) App {
	panic("not implemented")
}

func (m Mutation) CreateStaticTeamGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name    string
			TeamIDs []graphql.ID
		}
	},
) StaticTeamGroup {
	panic("not implemented")
}

func (m Mutation) UpdateStaticTeamGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name    string
			TeamIDs []graphql.ID
		}
	}) StaticTeamGroup {
	panic("not implemented")
}

func (m Mutation) CreateFilterTeamGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name   string
			Filter string
		}
	}) FilterTeamGroup {
	panic("not implemented")
}

func (m Mutation) UpdateFilterTeamGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name   string
			Filter string
		}
	},
) FilterTeamGroup {
	panic("not implemented")
}
