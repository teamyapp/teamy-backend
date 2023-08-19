package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Rollout interface {
	ID(ctx context.Context) graphql.ID
	Type(ctx context.Context) entity.RolloutType
	CreatedAt(ctx context.Context) graphql.Time
}

type UserRollout interface {
	Rollout
	UserGroups(ctx context.Context) []UserGroup
}

type TeamRollout interface {
	Rollout
	TeamGroups(ctx context.Context) []TeamGroup
}

type StaticRollout interface {
	Rollout
	AppVersion(ctx context.Context) AppVersion
}

type TimeRangeRollout interface {
	Rollout
	AppVersion(ctx context.Context) AppVersion
	StartAt(ctx context.Context) *graphql.Time
	EndAt(ctx context.Context) *graphql.Time
}

type ExperimentRollout interface {
	Rollout
	AppVersions(ctx context.Context) []AppVersion
	StartAt(ctx context.Context) *graphql.Time
	EndAt(ctx context.Context) *graphql.Time
}

func (m Mutation) DeleteRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
	}) Rollout {
	panic("implement me")
}
