package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskFilter struct {
	TaskID       *graphql.ID
	OwnerID      *graphql.ID
	GoalContains *string
	Status       *entity.TaskStatus
	IsPlanned    *bool
}

type TeamFilter struct {
	TeamID *graphql.ID
}

type InvitationFilter struct {
	InvitationID *graphql.ID
	Code         *string
}

type SprintFilter struct {
	SprintID        *graphql.ID
	StartAtAndAfter *graphql.Time
	SortByStartAt   *bool
	CountLimit      *int32
}

type AppFilter struct {
	IsPrivate *bool
}

func matchTeam(ct context.Context, dataCollector obs.DataCollector, filter TeamFilter, team entity.Team) bool {
	if filter.TeamID != nil {
		teamID, err := fromGraphQLIDPtr(ct, dataCollector, filter.TeamID)
		if err != nil {
			return false
		}

		if *teamID != team.ID {
			return false
		}
	}

	return true
}

func matchInvitation(ct context.Context, dataCollector obs.DataCollector, filter InvitationFilter, invitation entity.Invitation) bool {
	if filter.InvitationID != nil {
		invitationID, err := fromGraphQLIDPtr(ct, dataCollector, filter.InvitationID)
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
