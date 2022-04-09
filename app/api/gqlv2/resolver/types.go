package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"strconv"
	"time"
)

func toGraphQLID(id uint64) graphql.ID {
	return graphql.ID(strconv.FormatUint(id, 10))
}

func toGraphQLTime(time time.Time) graphql.Time {
	return graphql.Time{Time: time}
}
