package gql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

var availableStoryActions = map[entity.StoryStatus][]entity.StoryAction{
	entity.StoryStatusTodo: {
		entity.StoryActionStart,
		entity.StoryActionDelete,
		entity.StoryActionAssignOwner,
	},
	entity.StoryStatusInProgress: {
		entity.StoryActionMarkComplete,
		entity.StoryActionReportBlocked,
		entity.StoryActionAssignOwner,
		entity.StoryActionDelete,
	},
	entity.StoryStatusPaused: {
		entity.StoryActionStart,
		entity.StoryActionDelete,
		entity.StoryActionAssignOwner,
	},
	entity.StoryStatusDelivered: {
		entity.StoryActionDelete,
		entity.StoryActionAssignOwner,
	},
}

type Story struct {
	deps  *Dependencies
	story entity.Story
}

func (t Story) storyID() graphql.ID {
	return toGraphQLID(t.story.StoryID)
}

func (s Story) Goal() string {
	return s.story.Goal
}

func (s Story) Context() *string {
	return s.story.Context
}

func (s Story) Creator() (User, error) {
	user, err := s.deps.userService.FindUserByID(s.story.CreatorUserID)
	if err != nil {
		s.deps.logger.Error(err)
		return User{}, err
	}

	return User{deps: s.deps, user: user}, nil
}

func (s Story) Owner() (*User, error) {
	if s.story.OwnerUserID == nil {
		return nil, nil
	}

	user, err := s.deps.userService.FindUserByID(s.story.OwnerUserID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, errs.ToResolverErr(err)
	}

	return &User{deps: s.deps, user: user}, nil
}

func (s Story) OwningTeam() (Team, error) {
	team, err := s.deps.teamService.FindTeamByID(s.story.OwningTeamID)
	if err != nil {
		s.deps.logger.Error(err)
		return Team{}, errs.ToResolverErr(err)
	}

	return Team{deps: s.deps, team: team}, nil
}

func (s Story) Status() string {
	return string(s.story.Status)
}

func (s Story) IsPlanned() bool {
	return s.story.IsPlanned
}

func (s Story) Effort() *string {
	if s.story.Effort == nil {
		return nil
	}

	effort := s.story.Effort.String()
	return &effort
}

func (s Story) Priority() *string {
	if s.story.Priority == nil {
		return nil
	}

	priority := string(*s.story.Priority)
	return &priority
}

func (s Story) DueAt() *string {
	if s.story.DueAt == nil {
		return nil
	}

	dueAt := s.story.DueAt.Format("2006-01-02")
	return &dueAt
}

func (s Story) CommentsThread() (CommentsThread, error) {
	thread, err := s.deps.commentsService.FindCommentsThreadByID(s.story.CommentsThreadID)
	if err != nil {
		s.deps.logger.Error(err)
		return CommentsThread{}, err
	}

	return CommentsThread{deps: s.deps, thread: thread}, nil
}

func (s Story) CreatedAt() graphql.Time {
	return toGraphQLTime(s.story.CreatedAt)
}

func (s Story) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(s.story.UpdatedAt)
}

func (s Story) DeliveredAt() *graphql.Time {
	return toGraphQLTimePtr(s.story.DeliveredAt)
}

func (s Story) AvailableStoryActions() []entity.StoryAction {
	return availableStoryActions[s.story.Status]
}

func (s Story) Sprint() (*Sprint, error) {
	sprint, err := s.deps.sprintService.FindSprintByStoryID(s.story.ID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, err
	}

	if sprint == nil {
		return nil, nil
	}

	return &Sprint{deps: s.deps, sprint: *sprint}, nil
}

func (s Story) AwaitForStories() ([]Story, error) {
	stories, err := s.deps.storyService.FindAwaitForStories(s.story.ID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, err
	}

	return collect.Map(stories, func(story entity.Story, _ int) Story {
		return Story{deps: s.deps, story: story}
	}), nil
}

func (s Story) Links() ([]StoryLink, error) {
	links, err := s.deps.storyLinkService.FindLinksByStoryID(s.story.ID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, err
	}

	return collect.Map(links, func(storyLink entity.StoryLink, _ int) StoryLink {
		return StoryLink{deps: s.deps, storyLink: storyLink}
	}), nil
}

func (s Story) Tasks() ([]Task, error) {
	tasks, err := s.deps.taskService.FindTasksByStoryID(s.story.ID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return Task{deps: s.deps, task: task}
	}), nil
}

func (s Story) Participants() ([]User, error) {
	users, err := s.deps.userService.FindUsersByStoryID(s.story.ID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, err
	}

	return collect.Map(users, func(user entity.User, _ int) User {
		return User{deps: s.deps, user: user}
	}), nil
}

func (s Story) SprintTaskRelations() ([]SprintTaskRelation, error) {
	relations, err := s.deps.sprintTaskRelationService.FindSprintTaskRelationsByStoryID(s.story.ID)
	if err != nil {
		s.deps.logger.Error(err)
		return nil, err
	}

	return collect.Map(relations, func(relation entity.SprintTaskRelation, _ int) SprintTaskRelation {
		return SprintTaskRelation{deps: s.deps, relation: relation}
	}), nil
}

func newStory(deps *Dependencies, story entity.Story) Story {
	return Story{
		deps:  deps,
		story: story,
	}
}

type StoryLink struct {
	deps      *Dependencies
	storyLink entity.StoryLink
}

func (s StoryLink) ID() string {
	return toGraphQLID(s.storyLink.ID)
}
