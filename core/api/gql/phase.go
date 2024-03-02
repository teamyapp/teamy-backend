package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Phase struct {
	deps  *Dependencies
	phase entity.Phase
}

func (p Phase) ID() graphql.ID {
	return toGraphQLID(p.phase.ID)
}

func (p Phase) Creator(ct context.Context) (User, error) {
	user, err := p.deps.userService.FindUserByID(ct, p.phase.CreatorID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(p.deps, user), nil
}

func (p Phase) Name() string {
	return p.phase.Name
}

func (p Phase) Status() entity.PhaseStatus {
	return p.phase.Status
}

func (p Phase) ExpectedStartAt() graphql.Time {
	return toGraphQLTime(p.phase.ExpectedStartAt)
}

func (p Phase) ActualStartAt() *graphql.Time {
	return toGraphQLTimePtr(p.phase.ActualStartAt)
}

func (p Phase) ExpectedEndAt() graphql.Time {
	return toGraphQLTime(p.phase.ExpectedEndAt)
}

func (p Phase) ActualEndAt() *graphql.Time {
	return toGraphQLTimePtr(p.phase.ActualEndAt)
}

func (p Phase) CreatedAt() graphql.Time {
	return toGraphQLTime(p.phase.CreatedAt)
}

func (p Phase) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(p.phase.UpdatedAt)
}

func (p Phase) Stories(ct context.Context) ([]Story, error) {
	panic("not implemented")
}

func (m Mutation) CreatePhase(ct context.Context, args struct {
	ProjectID graphql.ID
	Input     struct {
		Name            string
		ExpectedStartAt graphql.Time
		ExpectedEndAt   graphql.Time
	}
}) (Phase, error) {
	panic("not implemented")
}

func (m Mutation) UpdatePhase(ct context.Context, args struct {
	PhaseID graphql.ID
	Input   struct {
		Name            string
		ExpectedStartAt graphql.Time
		ExpectedEndAt   graphql.Time
		ActualStartAt   *graphql.Time
		ActualEndAt     *graphql.Time
		Status          entity.PhaseStatus
	}
}) (Phase, error) {
	panic("not implemented")
}

func (m Mutation) DeletePhase(ct context.Context, args struct {
	PhaseID graphql.ID
}) (graphql.ID, error) {
	panic("not implemented")
}

func newPhase(deps *Dependencies, phase entity.Phase) Phase {
	return Phase{
		deps:  deps,
		phase: phase,
	}
}
