package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StaticTeamRollout struct {
}

var _ TeamRollout = (*StaticTeamRollout)(nil)
var _ StaticRollout = (*StaticTeamRollout)(nil)

func (r StaticTeamRollout) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (r StaticTeamRollout) Type(ctx context.Context) entity.RolloutType {
	panic("not implemented")
}

func (r StaticTeamRollout) AppVersion(ctx context.Context) AppVersion {
	panic("not implemented")
}

func (r StaticTeamRollout) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (r StaticTeamRollout) TeamGroups(ctx context.Context) []TeamGroup {
	panic("not implemented")
}

type TimeRangeTeamRollout struct {
}

var _ TeamRollout = (*TimeRangeTeamRollout)(nil)
var _ TimeRangeRollout = (*TimeRangeTeamRollout)(nil)

func (r TimeRangeTeamRollout) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (r TimeRangeTeamRollout) Type(ctx context.Context) entity.RolloutType {
	panic("not implemented")
}

func (r TimeRangeTeamRollout) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (r TimeRangeTeamRollout) AppVersion(ctx context.Context) AppVersion {
	panic("implement me")
}

func (r TimeRangeTeamRollout) StartAt(ctx context.Context) *graphql.Time {
	panic("implement me")
}

func (r TimeRangeTeamRollout) EndAt(ctx context.Context) *graphql.Time {
	panic("implement me")
}

func (r TimeRangeTeamRollout) TeamGroups(ctx context.Context) []TeamGroup {
	panic("not implemented")
}

type ExperimentTeamRollout struct {
}

var _ TeamRollout = (*ExperimentTeamRollout)(nil)
var _ ExperimentRollout = (*ExperimentTeamRollout)(nil)

func (r ExperimentTeamRollout) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (r ExperimentTeamRollout) Type(ctx context.Context) entity.RolloutType {
	panic("not implemented")
}

func (r ExperimentTeamRollout) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (r ExperimentTeamRollout) AppVersions(ctx context.Context) []AppVersion {
	panic("not implemented")
}

func (r ExperimentTeamRollout) StartAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (r ExperimentTeamRollout) EndAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (r ExperimentTeamRollout) TeamGroups(ctx context.Context) []TeamGroup {
	panic("not implemented")
}

func (m Mutation) CreateStaticTeamRollout(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumber int32
			TeamGroupIDs  []graphql.ID
		}
	},
) StaticTeamRollout {
	panic("not implemented")
}

func (m Mutation) UpdateStaticTeamRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			AppID         graphql.ID
			VersionNumber int32
			TeamGroupIDs  []graphql.ID
		}
	},
) StaticTeamRollout {
	panic("not implemented")
}

func (m Mutation) CreateTimeRangeTeamRollout(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumber int32
			StartAt       *graphql.Time
			EndAt         *graphql.Time
			TeamGroupIDs  []graphql.ID
		}
	},
) TimeRangeTeamRollout {
	panic("not implemented")
}

func (m Mutation) UpdateTimeRangeTeamRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			AppID         graphql.ID
			VersionNumber int32
			StartAt       *graphql.Time
			EndAt         *graphql.Time
			TeamGroupIDs  []graphql.ID
		}
	},
) TimeRangeTeamRollout {
	panic("not implemented")
}

func (m Mutation) CreateExperimentTeamRollout(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumbers []int32
			StartAt        *graphql.Time
			EndAt          *graphql.Time
			TeamGroupIDs   []graphql.ID
		}
	},
) ExperimentTeamRollout {
	panic("not implemented")
}

func (m Mutation) UpdateExperimentTeamRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			AppID          graphql.ID
			VersionNumbers []int32
			StartAt        *graphql.Time
			EndAt          *graphql.Time
			TeamGroupIDs   []graphql.ID
		}
	},
) ExperimentTeamRollout {
	panic("not implemented")
}
