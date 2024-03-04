package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Story struct {
	deps  *Dependencies
	story entity.Story
}

func (s Story) ID() graphql.ID {
	return toGraphQLID(s.story.ID)
}

func (s Story) Name() string {
	return s.story.Name
}

func (s Story) Creator(ct context.Context) (User, error) {
	user, err := s.deps.userService.FindUserByID(ct, s.story.CreatorID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(s.deps, user), nil
}

func (s Story) Owner(ct context.Context) (User, error) {
	user, err := s.deps.userService.FindUserByID(ct, s.story.OwnerID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(s.deps, user), nil
}

func (s Story) Status() entity.StoryStatus {
	return s.story.Status
}

func (s Story) Priority() entity.Priority {
	return s.story.Priority
}

func (s Story) CreatedAt() graphql.Time {
	return toGraphQLTime(s.story.CreatedAt)
}

func (s Story) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(s.story.UpdatedAt)
}

func (s Story) Tasks(ct context.Context) ([]Task, error) {
	tasks, err := s.deps.storyService.FindTasksByStoryID(ct, s.story.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(s.deps, task)
	}), nil
}

func (m Mutation) CreateStory(ct context.Context, args struct {
	ProjectID graphql.ID
	Input     struct {
		Name     string
		OwnerID  graphql.ID
		Priority entity.Priority
	}
}) (Story, error) {
	projectID, internalErr := fromGraphQLID(args.ProjectID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	ownerID, internalErr := fromGraphQLID(args.Input.OwnerID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	createStoryInput := service.CreateStoryInput{
		Name:     args.Input.Name,
		OwnerID:  ownerID,
		Priority: args.Input.Priority,
	}

	story, err := m.deps.storyService.CreateStory(ct, projectID, createStoryInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func (m Mutation) UpdateStory(ct context.Context, args struct {
	StoryID graphql.ID
	Input   struct {
		Name     string
		OwnerID  graphql.ID
		Status   entity.StoryStatus
		Priority entity.Priority
	}
}) (Story, error) {
	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	ownerID, internalErr := fromGraphQLID(args.Input.OwnerID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	updateStoryInput := service.UpdateStoryInput{
		Name:     args.Input.Name,
		OwnerID:  ownerID,
		Status:   args.Input.Status,
		Priority: args.Input.Priority,
	}

	story, err := m.deps.storyService.UpdateStory(ct, storyID, updateStoryInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func (m Mutation) DeleteStory(ct context.Context, args struct {
	StoryID graphql.ID
}) (Story, error) {
	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	story, err := m.deps.storyService.DeleteStory(ct, storyID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func (m Mutation) AddTaskToStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskID  graphql.ID
}) (Story, error) {
	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	taskID, internalErr := fromGraphQLID(args.TaskID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	story, err := m.deps.storyService.AddTaskToStory(ct, storyID, taskID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func (m Mutation) AddTasksToStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskIDs []graphql.ID
}) (Story, error) {
	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	taskIDs, err := collect.MapWithErr(args.TaskIDs, func(taskID graphql.ID, _ int) (uint64, *errs.Error) {
		id, internalErr := fromGraphQLID(taskID)
		if internalErr != nil {
			return 0, errs.NewError(
				errs.InvalidArgument,
				internalErr.Error(),
			)
		}

		return id, nil
	})

	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	story, err := m.deps.storyService.AddTasksToStory(ct, storyID, taskIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func (m Mutation) RemoveTaskFromStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskID  graphql.ID
}) (Story, error) {
	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	taskID, internalErr := fromGraphQLID(args.TaskID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	story, err := m.deps.storyService.RemoveTaskFromStory(ct, storyID, taskID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func (m Mutation) RemoveTasksFromStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskIDs []graphql.ID
}) (Story, error) {
	storyID, internalErr := fromGraphQLID(args.StoryID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Story{}, errs.ToResolverErr(internalErr)
	}

	taskIDs, err := collect.MapWithErr(args.TaskIDs, func(taskID graphql.ID, _ int) (uint64, *errs.Error) {
		id, internalErr := fromGraphQLID(taskID)
		if internalErr != nil {
			return 0, errs.NewError(
				errs.InvalidArgument,
				internalErr.Error(),
			)
		}

		return id, nil
	})

	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	story, err := m.deps.storyService.RemoveTasksFromStory(ct, storyID, taskIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Story{}, errs.ToResolverErr(err)
	}

	return newStory(m.deps, story), nil
}

func newStory(deps *Dependencies, story entity.Story) Story {
	return Story{deps, story}
}
