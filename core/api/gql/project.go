package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Project struct {
	deps    *Dependencies
	project entity.Project
}

func (p Project) ID() graphql.ID {
	return toGraphQLID(p.project.ID)
}

func (p Project) Creator(ct context.Context) (User, error) {
	user, err := p.deps.userService.FindUserByID(ct, p.project.CreatorID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(p.deps, user), nil
}

func (p Project) Name() string {
	return p.project.Name
}

func (p Project) ExpectedStartAt() *graphql.Time {
	return toGraphQLTimePtr(p.project.ExpectedStartAt)
}

func (p Project) ActualStartAt() *graphql.Time {
	return toGraphQLTimePtr(p.project.ActualStartAt)
}

func (p Project) ExpectedEndAt() *graphql.Time {
	return toGraphQLTimePtr(p.project.ExpectedEndAt)
}

func (p Project) ActualEndAt() *graphql.Time {
	return toGraphQLTimePtr(p.project.ActualEndAt)
}

func (p Project) CreatedAt() graphql.Time {
	return toGraphQLTime(p.project.CreatedAt)
}

func (p Project) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(p.project.UpdatedAt)
}

func (p Project) Phases(ct context.Context) ([]Phase, error) {
	panic("not implemented")
}

func (p Project) Stories(ct context.Context) ([]Story, error) {
	panic("not implemented")
}

func (m Mutation) CreateProject(ct context.Context, args struct {
	Input struct {
		Name            string
		ExpectedStartAt *graphql.Time
		ExpectedEndAt   *graphql.Time
	}
}) (Project, error) {
	panic("not implemented")
}

func (m Mutation) UpdateProject(ct context.Context, args struct {
	ProjectID graphql.ID
	Input     struct {
		Name            string
		ExpectedStartAt *graphql.Time
		ActualStartAt   *graphql.Time
		ExpectedEndAt   *graphql.Time
		ActualEndAt     *graphql.Time
	}
}) (Project, error) {
	panic("not implemented")
}

func (m Mutation) DeleteProject(ct context.Context, args struct {
	ProjectID graphql.ID
}) (Project, error) {
	panic("not implemented")
}

func newProject(deps *Dependencies, project entity.Project) Project {
	return Project{
		deps:    deps,
		project: project,
	}
}
