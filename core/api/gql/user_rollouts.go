package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StaticUserRollout struct {
}

var _ UserRollout = (*StaticUserRollout)(nil)
var _ StaticRollout = (*StaticUserRollout)(nil)

func (r StaticUserRollout) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (r StaticUserRollout) Type(ctx context.Context) entity.RolloutType {
	panic("not implemented")
}

func (r StaticUserRollout) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (r StaticUserRollout) AppVersion(ctx context.Context) AppVersion {
	panic("implement me")
}

func (r TimeRangeUserRollout) StartAt(ctx context.Context) *graphql.Time {
	panic("implement me")
}

func (r TimeRangeUserRollout) EndAt(ctx context.Context) *graphql.Time {
	panic("implement me")
}

func (r StaticUserRollout) UserGroups(ctx context.Context) []UserGroup {
	panic("not implemented")
}

type TimeRangeUserRollout struct {
}

var _ UserRollout = (*TimeRangeUserRollout)(nil)
var _ TimeRangeRollout = (*TimeRangeUserRollout)(nil)

func (r TimeRangeUserRollout) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (r TimeRangeUserRollout) Type(ctx context.Context) entity.RolloutType {
	panic("not implemented")
}

func (r TimeRangeUserRollout) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (r TimeRangeUserRollout) UserGroups(ctx context.Context) []UserGroup {
	panic("not implemented")
}

func (r TimeRangeUserRollout) AppVersion(ctx context.Context) AppVersion {
	panic("implement me")
}

type ExperimentUserRollout struct {
}

var _ UserRollout = (*ExperimentUserRollout)(nil)
var _ ExperimentRollout = (*ExperimentUserRollout)(nil)

func (r ExperimentUserRollout) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (r ExperimentUserRollout) Type(ctx context.Context) entity.RolloutType {
	panic("not implemented")
}

func (r ExperimentUserRollout) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (r ExperimentUserRollout) AppVersions(ctx context.Context) []AppVersion {
	panic("implement me")
}

func (r ExperimentUserRollout) StartAt(ctx context.Context) *graphql.Time {
	panic("implement me")
}

func (r ExperimentUserRollout) EndAt(ctx context.Context) *graphql.Time {
	panic("implement me")
}

func (r ExperimentUserRollout) UserGroups(ctx context.Context) []UserGroup {
	panic("not implemented")
}

func (m Mutation) CreateStaticUserRollout(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumber int32
			UserGroupIds  []graphql.ID
		}
	},
) (StaticUserRollout, error) {
	panic("implement me")
}

func (m Mutation) UpdateStaticUserRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			AppID         graphql.ID
			VersionNumber int32
			UserGroupIds  []graphql.ID
		}
	},
) (StaticUserRollout, error) {
	panic("implement me")
}

func (m Mutation) CreateTimeRangeUserRollout(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumber int32
			StartAt       *graphql.Time
			EndAt         *graphql.Time
			UserGroupIds  []graphql.ID
		}
	},
) (TimeRangeUserRollout, error) {
	panic("implement me")
}

func (m Mutation) UpdateTimeRangeUserRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			AppID         graphql.ID
			VersionNumber int32
			StartAt       *graphql.Time
			EndAt         *graphql.Time
			UserGroupIds  []graphql.ID
		}
	},
) (TimeRangeUserRollout, error) {
	panic("implement me")
}

func (m Mutation) CreateExperimentUserRollout(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumbers []int32
			StartAt        *graphql.Time
			EndAt          *graphql.Time
			UserGroupIds   []graphql.ID
		}
	},
) (ExperimentUserRollout, error) {
	panic("implement me")
}

func (m Mutation) UpdateExperimentUserRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			AppID          graphql.ID
			VersionNumbers []int32
			StartAt        *graphql.Time
			EndAt          *graphql.Time
			UserGroupIds   []graphql.ID
		}
	},
) (ExperimentUserRollout, error) {
	panic("implement me")
}
