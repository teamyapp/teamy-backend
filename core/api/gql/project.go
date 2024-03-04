package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Project struct {
	deps    *Dependencies
	project entity.Project
}

func (p Project) ID() graphql.ID {
	return toGraphQLID(p.project.ID)
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

func (p Project) Creator(ct context.Context) (User, error) {
	user, err := p.deps.userService.FindUserByID(ct, p.project.CreatorID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(p.deps, user), nil
}

func (p Project) CreatedAt() graphql.Time {
	return toGraphQLTime(p.project.CreatedAt)
}

func (p Project) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(p.project.UpdatedAt)
}

func (p Project) Phases(ct context.Context) ([]Phase, error) {
	phases, err := p.deps.projectService.FindPhasesByProjectID(ct, p.project.ID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(phases, func(phase entity.Phase, _ int) Phase {
		return newPhase(p.deps, phase)
	}), nil
}

func (p Project) Stories(ct context.Context) ([]Story, error) {
	stories, err := p.deps.projectService.FindStoriesByProjectID(ct, p.project.ID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(stories, func(story entity.Story, _ int) Story {
		return newStory(p.deps, story)
	}), nil
}

func (m Mutation) CreateProject(ct context.Context, args struct {
	Input struct {
		Name            string
		ExpectedStartAt *graphql.Time
		ExpectedEndAt   *graphql.Time
	}
}) (Project, error) {
	createProjectInput := service.CreateProjectInput{
		Name:            args.Input.Name,
		ExpectedStartAt: fromGraphQLTimePtr(args.Input.ExpectedStartAt),
		ExpectedEndAt:   fromGraphQLTimePtr(args.Input.ExpectedEndAt),
	}

	project, err := m.deps.projectService.CreateProject(ct, createProjectInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Project{}, errs.ToResolverErr(err)
	}

	return newProject(m.deps, project), nil
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
	projectID, internalErr := fromGraphQLID(args.ProjectID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Project{}, errs.ToResolverErr(internalErr)
	}

	updateProjectInput := service.UpdateProjectInput{
		Name:            args.Input.Name,
		ExpectedStartAt: fromGraphQLTimePtr(args.Input.ExpectedStartAt),
		ActualStartAt:   fromGraphQLTimePtr(args.Input.ActualStartAt),
		ExpectedEndAt:   fromGraphQLTimePtr(args.Input.ExpectedEndAt),
		ActualEndAt:     fromGraphQLTimePtr(args.Input.ActualEndAt),
	}

	project, err := m.deps.projectService.UpdateProject(ct, projectID, updateProjectInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Project{}, errs.ToResolverErr(err)
	}

	return newProject(m.deps, project), nil
}

func (m Mutation) DeleteProject(ct context.Context, args struct {
	ProjectID graphql.ID
}) (Project, error) {
	projectID, internalErr := fromGraphQLID(args.ProjectID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Project{}, errs.ToResolverErr(internalErr)
	}

	project, err := m.deps.projectService.DeleteProject(ct, projectID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Project{}, errs.ToResolverErr(err)
	}

	return newProject(m.deps, project), nil
}

func newProject(deps *Dependencies, project entity.Project) Project {
	return Project{
		deps:    deps,
		project: project,
	}
}
