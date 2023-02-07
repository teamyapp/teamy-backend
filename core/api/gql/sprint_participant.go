package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant struct {
	deps        *Dependencies
	participant entity.SprintParticipant
}

func (s SprintParticipant) Sprint(ct context.Context) (Sprint, error) {
	sprint, err := s.deps.sprintService.FindSprintByID(ct, s.participant.SprintID)
	if err != nil {
		s.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return Sprint{}, errs.ToResolverErr(err)
	}

	return newSprint(s.deps, sprint), nil
}

func (s SprintParticipant) User(ct context.Context) (User, error) {
	user, err := s.deps.userService.FindUserByID(ct, s.participant.UserID)
	if err != nil {
		s.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(s.deps, user), nil
}

func (s SprintParticipant) TotalBandwidth(ct context.Context) scalar.Duration {
	return scalar.Duration{Duration: s.participant.TotalBandwidth}
}

func (s SprintParticipant) UnusedBandwidth(ct context.Context) scalar.Duration {
	return scalar.Duration{Duration: s.participant.UnusedBandwidth}
}

func (s SprintParticipant) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.participant.CreatedAt)
}

func (s SprintParticipant) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(s.participant.UpdatedAt)
}

func newSprintParticipant(
	deps *Dependencies,
	participant entity.SprintParticipant,
) SprintParticipant {
	return SprintParticipant{
		deps:        deps,
		participant: participant,
	}
}
