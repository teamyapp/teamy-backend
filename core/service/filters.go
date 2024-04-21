package service

import (
	"slices"
	"strings"
	"time"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskFilter struct {
	TaskID       *uint64
	OwnerID      *uint64
	GoalContains *string
	Status       *entity.TaskStatus
	IsScheduled  *bool
	IsPlanned    *bool
}

type SprintFilter struct {
	SprintID        *uint64
	StartAtAndAfter *time.Time
	SortByStartAt   *bool
	CountLimit      *int
}

type TeamFilter struct {
	TeamID *uint64
}

type ProjectFilter struct {
	ProjectID *uint64
}

type PhaseFilter struct {
	PhaseID *uint64
}

type StoryFilter struct {
	StoryID *uint64
}

type InvitationFilter struct {
	InvitationID *uint64
	Code         *string
}

func filterTasks(tasks []entity.Task, filter TaskFilter) []entity.Task {
	return collect.Filter(tasks, func(task entity.Task) bool {
		if filter.TaskID != nil && *filter.TaskID != task.ID {
			return false
		}

		if filter.OwnerID != nil {
			if task.OwnerUserID == nil || *filter.OwnerID != *task.OwnerUserID {
				return false
			}
		}

		if filter.Status != nil && *filter.Status != task.Status {
			return false
		}

		if filter.IsScheduled != nil && *filter.IsScheduled != task.IsScheduled {
			return false
		}

		if filter.IsPlanned != nil && *filter.IsPlanned != task.IsPlanned {
			return false
		}

		if filter.GoalContains != nil &&
			!strings.Contains(strings.ToLower(task.Goal), strings.ToLower(*filter.GoalContains)) {
			return false
		}

		return true
	})
}

func filterSprints(sprints []entity.Sprint, filter SprintFilter) []entity.Sprint {
	sprints = collect.Filter(sprints, func(sprint entity.Sprint) bool {
		if filter.SprintID != nil && *filter.SprintID != sprint.ID {
			return false
		}

		if filter.StartAtAndAfter != nil && (*filter.StartAtAndAfter).After(sprint.StartAt) {
			return false
		}

		return true
	})

	if filter.SortByStartAt != nil {
		slices.SortStableFunc(sprints, func(sprint1 entity.Sprint, sprint2 entity.Sprint) int {
			if sprint1.StartAt.Before(sprint2.StartAt) {
				return -1
			} else if sprint1.StartAt.After(sprint2.StartAt) {
				return 1
			}

			return 0
		})
	}

	if filter.CountLimit != nil && *filter.CountLimit > len(sprints) {
		sprints = sprints[:*filter.CountLimit]
	}

	return sprints
}

func filterTeams(teams []entity.Team, filter TeamFilter) []entity.Team {
	return collect.Filter(teams, func(team entity.Team) bool {
		if filter.TeamID != nil && *filter.TeamID != team.ID {
			return false
		}

		return true
	})
}

func filterProjects(projects []entity.Project, filter ProjectFilter) []entity.Project {
	return collect.Filter(projects, func(project entity.Project) bool {
		if filter.ProjectID != nil && *filter.ProjectID != project.ID {
			return false
		}

		return true
	})
}

func filterPhases(phases []entity.Phase, filter PhaseFilter) []entity.Phase {
	return collect.Filter(phases, func(phase entity.Phase) bool {
		if filter.PhaseID != nil && *filter.PhaseID != phase.ID {
			return false
		}

		return true
	})
}

func filterStories(stories []entity.Story, filter StoryFilter) []entity.Story {
	return collect.Filter(stories, func(story entity.Story) bool {
		if filter.StoryID != nil && *filter.StoryID != story.ID {
			return false
		}

		return true
	})
}

func filterInvitations(invitations []entity.Invitation, filter InvitationFilter) []entity.Invitation {
	return collect.Filter(invitations, func(invitation entity.Invitation) bool {
		if filter.InvitationID != nil && *filter.InvitationID != invitation.ID {
			return false
		}

		if filter.Code != nil && *filter.Code != invitation.Code {
			return false
		}

		return true
	})
}
