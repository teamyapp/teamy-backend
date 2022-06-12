package resolver

import (
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskFilter struct {
	TaskID       *graphql.ID
	OwnerID      *graphql.ID
	GoalContains *string
	Status       *entity.TaskStatus
}

type TeamFilter struct {
	TeamID *graphql.ID
}

type InvitationFilter struct {
	InvitationID *graphql.ID
	Code         *string
}

func matchTask(filter TaskFilter, task entity.Task) bool {
	if filter.TaskID != nil {
		taskID, err := fromGraphQLIDPtr(filter.TaskID)
		if err != nil {
			return false
		}

		if *taskID != task.ID {
			return false
		}
	}

	if filter.OwnerID != nil {
		ownerID, err := fromGraphQLIDPtr(filter.OwnerID)
		if err != nil {
			return false
		}

		if task.OwnerUserID == nil || *ownerID != *task.OwnerUserID {
			return false
		}
	}

	if filter.Status != nil && *filter.Status != task.Status {
		return false
	}

	if filter.GoalContains != nil &&
		!strings.Contains(strings.ToLower(task.Goal), strings.ToLower(*filter.GoalContains)) {
		return false
	}

	return true
}

func matchTeam(filter TeamFilter, team entity.Team) bool {
	if filter.TeamID != nil {
		teamID, err := fromGraphQLIDPtr(filter.TeamID)
		if err != nil {
			return false
		}

		if *teamID != team.ID {
			return false
		}
	}

	return true
}

func matchInvitation(filter InvitationFilter, invitation entity.Invitation) bool {
	if filter.InvitationID != nil {
		invitationID, err := fromGraphQLIDPtr(filter.InvitationID)
		if err != nil {
			return false
		}

		if *invitationID != invitation.ID {
			return false
		}
	}

	if filter.Code != nil && *filter.Code != invitation.Code {
		return false
	}

	return true
}
