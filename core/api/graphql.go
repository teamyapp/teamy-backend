package api

import (
	_ "embed"

	"github.com/graph-gophers/graphql-go/trace/tracer"
	cloudGQL "github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/gql"
)

//go:embed gql/schema.graphql
var rawSchema string

func NewGraphQL(dataCollector telemetry.DataCollector, tracer tracer.Tracer, resolver gql.Resolver) cloudGQL.Service[gql.Resolver] {
	return cloudGQL.NewService(dataCollector, tracer, rawSchema, &resolver, graphQLPrefix)
}
