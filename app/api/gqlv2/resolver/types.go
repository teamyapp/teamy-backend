package resolver

import (
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
)

func toGraphQLID(id uint64) graphql.ID {
	return graphql.ID(strconv.FormatUint(id, 10))
}

func toGraphQLTime(time time.Time) graphql.Time {
	return graphql.Time{Time: time}
}

func toGraphQLTimePtr(time *time.Time) *graphql.Time {
	return &graphql.Time{Time: *time}
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

func fromGraphQLID(graphqlID graphql.ID) (uint64, error) {
	return strconv.ParseUint(string(graphqlID), 10, 64)
}
