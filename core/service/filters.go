package service

import (
	"sort"
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
		sort.Slice(sprints, func(i, j int) bool {
			return sprints[i].StartAt.Before(sprints[j].StartAt)
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
