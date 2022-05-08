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

func fromGraphQLID(graphqlID graphql.ID) (uint64, error) {
	return strconv.ParseUint(string(graphqlID), 10, 64)
}
