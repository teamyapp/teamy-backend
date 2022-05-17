package resolver

import (
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type TaskFilter struct {
	TaskID  *graphql.ID
	OwnerID *graphql.ID
	Goal    *string
	Status  *entityv2.TaskStatus
}

type TeamFilter struct {
	TeamID *graphql.ID
}

type InvitationFilter struct {
	InvitationID *graphql.ID
	Code         *string
}

func matchTask(filter TaskFilter, task entityv2.Task) bool {
	if filter.OwnerID != nil {
		ownerID, err := fromGraphQLIDPtr(filter.OwnerID)
		if err != nil {
			return false
		}

		if ownerID != task.OwnerUserID {
			return false
		}
	}

	if filter.Status != nil && *filter.Status != task.Status {
		return false
	}

	if filter.Goal != nil && !strings.Contains(task.Goal, *filter.Goal) {
		return false
	}

	return true
}

func matchInvitation(filter InvitationFilter, invitation entityv2.Invitation) bool {
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
