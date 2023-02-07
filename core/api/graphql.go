package api

import (
	_ "embed"

	cloudGQL "github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/gql"
)

//go:embed gql/schema.graphql
var rawSchema string

func NewGraphQL(dataCollector telemetry.DataCollector, resolver gql.Resolver) cloudGQL.Service[gql.Resolver] {
	return cloudGQL.NewService(dataCollector, rawSchema, &resolver, graphQLPrefix)
}
