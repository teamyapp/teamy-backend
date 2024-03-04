package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Phase struct {
	deps  *Dependencies
	phase entity.Phase
}

func (p Phase) ID() graphql.ID {
	return toGraphQLID(p.phase.ID)
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

func (p Phase) Creator(ct context.Context) (User, error) {
	user, err := p.deps.userService.FindUserByID(ct, p.phase.CreatorID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(p.deps, user), nil
}

func (p Phase) CreatedAt() graphql.Time {
	return toGraphQLTime(p.phase.CreatedAt)
}

func (p Phase) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(p.phase.UpdatedAt)
}

func (p Phase) Stories(ct context.Context) ([]Story, error) {
	stories, err := p.deps.phaseService.FindStoriesByPhaseID(ct, p.phase.ID)
	if err != nil {
		p.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(stories, func(story entity.Story, _ int) Story {
		return newStory(p.deps, story)
	}), nil
}

func (m Mutation) CreatePhase(ct context.Context, args struct {
	ProjectID graphql.ID
	Input     struct {
		Name            string
		ExpectedStartAt graphql.Time
		ExpectedEndAt   graphql.Time
	}
}) (Phase, error) {
	projectID, internalErr := fromGraphQLID(args.ProjectID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	createPhaseInput := service.CreatePhaseInput{
		Name:            args.Input.Name,
		ExpectedStartAt: fromGraphQLTime(args.Input.ExpectedStartAt),
		ExpectedEndAt:   fromGraphQLTime(args.Input.ExpectedEndAt),
	}

	phase, err := m.deps.phaseService.CreatePhase(ct, projectID, createPhaseInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
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
	phaseID, internalErr := fromGraphQLID(args.PhaseID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	updatePhaseInput := service.UpdatePhaseInput{
		Name:            args.Input.Name,
		ExpectedStartAt: fromGraphQLTime(args.Input.ExpectedStartAt),
		ExpectedEndAt:   fromGraphQLTime(args.Input.ExpectedEndAt),
		ActualStartAt:   fromGraphQLTimePtr(args.Input.ActualStartAt),
		ActualEndAt:     fromGraphQLTimePtr(args.Input.ActualEndAt),
		Status:          args.Input.Status,
	}

	phase, err := m.deps.phaseService.UpdatePhase(ct, phaseID, updatePhaseInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
}

func (m Mutation) DeletePhase(ct context.Context, args struct {
	PhaseID graphql.ID
}) (Phase, error) {
	phaseID, internalErr := fromGraphQLID(args.PhaseID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	phase, err := m.deps.phaseService.DeletePhase(ct, phaseID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
}

func (m Mutation) AddStoryToPhase(ct context.Context, args struct {
	PhaseID graphql.ID
	StoryID graphql.ID
}) (Phase, error) {
	phaseID, internalErr := fromGraphQLID(args.PhaseID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	phase, err := m.deps.phaseService.AddStoryToPhase(ct, phaseID, storyID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
}

func (m Mutation) AddStoriesToPhase(ct context.Context, args struct {
	PhaseID  graphql.ID
	StoryIDs []graphql.ID
}) (Phase, error) {
	phaseID, internalErr := fromGraphQLID(args.PhaseID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	storyIDs, err := collect.MapWithErr(
		args.StoryIDs,
		func(storyID graphql.ID, _ int) (uint64, *errs.Error) {
			id, internalErr := fromGraphQLID(storyID)
			if internalErr != nil {
				return 0, errs.NewError(
					errs.InvalidArgument,
					internalErr.Error(),
				)
			}

			return id, nil
		},
	)

	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	phase, err := m.deps.phaseService.AddStoriesToPhase(ct, phaseID, storyIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
}

func (m Mutation) RemoveStoryFromPhase(ct context.Context, args struct {
	PhaseID graphql.ID
	StoryID graphql.ID
}) (Phase, error) {
	phaseID, internalErr := fromGraphQLID(args.PhaseID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	phase, err := m.deps.phaseService.RemoveStoryFromPhase(ct, phaseID, storyID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
}

func (m Mutation) RemoveStoriesFromPhase(ct context.Context, args struct {
	PhaseID  graphql.ID
	StoryIDs []graphql.ID
}) (Phase, error) {
	phaseID, internalErr := fromGraphQLID(args.PhaseID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Phase{}, errs.ToResolverErr(internalErr)
	}

	storyIDs, err := collect.MapWithErr(
		args.StoryIDs,
		func(storyID graphql.ID, _ int) (uint64, *errs.Error) {
			id, internalErr := fromGraphQLID(storyID)
			if internalErr != nil {
				return 0, errs.NewError(
					errs.InvalidArgument,
					internalErr.Error(),
				)
			}

			return id, nil
		})

	phase, err := m.deps.phaseService.RemoveStoriesFromPhase(ct, phaseID, storyIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Phase{}, errs.ToResolverErr(err)
	}

	return newPhase(m.deps, phase), nil
}

func newPhase(deps *Dependencies, phase entity.Phase) Phase {
	return Phase{
		deps:  deps,
		phase: phase,
	}
}
