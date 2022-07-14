package gql

import (
	"log"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/service"
)

func toGraphQLID(id uint64) graphql.ID {
	return graphql.ID(strconv.FormatUint(id, 10))
}

func toGraphQLTime(tm time.Time) graphql.Time {
	return graphql.Time{Time: tm}
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
		log.Println(err)
		return nil, err
	}

	ownerID, err := fromGraphQLIDPtr(gqlTaskFilter.OwnerID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &service.TaskFilter{
		TaskID:       taskID,
		OwnerID:      ownerID,
		GoalContains: gqlTaskFilter.GoalContains,
		Status:       gqlTaskFilter.Status,
	}, nil
}

func fromGraphQLSprintFilterPtr(gqlSprintFilter *SprintFilter) (*service.SprintFilter, error) {
	if gqlSprintFilter == nil {
		return nil, nil
	}

	sprintID, err := fromGraphQLIDPtr(gqlSprintFilter.SprintID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &service.SprintFilter{
		SprintID:        sprintID,
		StartAtAndAfter: fromGraphQLTimePtr(gqlSprintFilter.StartAtAndAfter),
		SortByStartAt:   gqlSprintFilter.SortByStartAt,
		CountLimit:      intPtrFromIntPtr(gqlSprintFilter.CountLimit),
	}, nil
}
