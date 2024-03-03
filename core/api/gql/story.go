package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
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
	panic("not implemented")
}

func (m Mutation) CreateStory(ct context.Context, args struct {
	ProjectID graphql.ID
	Input     struct {
		Name     string
		OwnerID  graphql.ID
		Priority entity.Priority
	}
}) (Story, error) {
	panic("not implemented")
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
	panic("not implemented")
}

func (m Mutation) DeleteStory(ct context.Context, args struct {
	StoryID graphql.ID
}) (Story, error) {
	panic("not implemented")
}

func (m Mutation) AddTaskToStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskID  graphql.ID
}) (Story, error) {
	panic("not implemented")
}

func (m Mutation) AddTasksToStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskIDs []graphql.ID
}) (Story, error) {
	panic("not implemented")
}

func (m Mutation) RemoveTaskFromStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskID  graphql.ID
}) (Story, error) {
	panic("not implemented")
}

func (m Mutation) RemoveTasksFromStory(ct context.Context, args struct {
	StoryID graphql.ID
	TaskIDs []graphql.ID
}) (Story, error) {
	panic("not implemented")
}

func newStory(deps *Dependencies, story entity.Story) Story {
	return Story{deps, story}
}
