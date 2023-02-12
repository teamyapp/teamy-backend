package gql

import (
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/service"
)

func toGraphQLID(id uint64) graphql.ID {
	return graphql.ID(strconv.FormatUint(id, 10))
}

func toGraphQLTime(tm time.Time) graphql.Time {
	return graphql.Time{Time: tm}
}

func toGraphQLDuration(duration time.Duration) scalar.Duration {
	return scalar.Duration{Duration: duration}
}

func toGraphQLTimePtr(tm *time.Time) *graphql.Time {
	if tm == nil {
		return nil
	}

	return &graphql.Time{Time: *tm}
}

func fromGraphQLTimePtr(tm *graphql.Time) *time.Time {
	if tm == nil {
		return nil
	}

	return &tm.Time
}

func toGraphQLDurationPtr(du *time.Duration) *scalar.Duration {
	if du == nil {
		return nil
	}

	return &scalar.Duration{Duration: *du}
}

func fromGraphQLIDPtr(graphqlID *graphql.ID) (*uint64, error) {
	if graphqlID == nil {
		return nil, nil
	}

	id, err := fromGraphQLID(*graphqlID)
	if err != nil {
		return nil, err
	}

	return &id, err
}

func intPtrFromIntPtr(num *int32) *int {
	if num == nil {
		return nil
	}

	newNum := int(*num)
	return &newNum
}

func int32PtrFromIntPtr(num *int) *int32 {
	if num == nil {
		return nil
	}

	newNum := int32(*num)
	return &newNum
}

func fromGraphQLID(graphqlID graphql.ID) (uint64, error) {
	return strconv.ParseUint(string(graphqlID), 10, 64)
}

func fromGraphQLTaskFilterPtr(gqlTaskFilter *TaskFilter) (*service.TaskFilter, error) {
	if gqlTaskFilter == nil {
		return nil, nil
	}

	taskID, err := fromGraphQLIDPtr(gqlTaskFilter.TaskID)
	if err != nil {
		return nil, err
	}

	ownerID, err := fromGraphQLIDPtr(gqlTaskFilter.OwnerID)
	if err != nil {
		return nil, err
	}

	return &service.TaskFilter{
		TaskID:       taskID,
		OwnerID:      ownerID,
		GoalContains: gqlTaskFilter.GoalContains,
		Status:       gqlTaskFilter.Status,
		IsPlanned:    gqlTaskFilter.IsPlanned,
	}, nil
}

func fromGraphQLSprintFilterPtr(gqlSprintFilter *SprintFilter) (*service.SprintFilter, error) {
	if gqlSprintFilter == nil {
		return nil, nil
	}

	sprintID, err := fromGraphQLIDPtr(gqlSprintFilter.SprintID)
	if err != nil {
		return nil, err
	}

	return &service.SprintFilter{
		SprintID:        sprintID,
		StartAtAndAfter: fromGraphQLTimePtr(gqlSprintFilter.StartAtAndAfter),
		SortByStartAt:   gqlSprintFilter.SortByStartAt,
		CountLimit:      intPtrFromIntPtr(gqlSprintFilter.CountLimit),
	}, nil
}

func fromGraphQLTeamFilterPtr(teamFilter *TeamFilter) (*service.TeamFilter, error) {
	if teamFilter == nil {
		return nil, nil
	}

	teamID, err := fromGraphQLIDPtr(teamFilter.TeamID)
	if err != nil {
		return nil, err
	}

	return &service.TeamFilter{
		TeamID: teamID,
	}, nil
}

func fromGraphQLInvitationFilterPtr(filter *InvitationFilter) (*service.InvitationFilter, error) {
	if filter == nil {
		return nil, nil
	}

	invitationID, err := fromGraphQLIDPtr(filter.InvitationID)
	if err != nil {
		return nil, err
	}

	return &service.InvitationFilter{
		InvitationID: invitationID,
		Code:         filter.Code,
	}, nil
}

func fromGraphQLAppFilterPtr(filter *AppFilter) (*service.AppFilter, error) {
	if filter == nil {
		return nil, nil
	}

	appID, err := fromGraphQLIDPtr(filter.AppID)
	if err != nil {
		return nil, err
	}

	teamID, err := fromGraphQLIDPtr(filter.TeamID)
	if err != nil {
		return nil, err
	}

	return &service.AppFilter{
		AppID:  appID,
		TeamID: teamID,
	}, nil
}

func fromGraphQLDurationPtr(duration *scalar.Duration) *time.Duration {
	if duration == nil {
		return nil
	}

	return &duration.Duration
}
