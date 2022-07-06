package api

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/core/api/gql"
)

//go:embed gql/schema.graphql
var rawSchema string

type GraphQL struct {
	gqlResolver gql.Resolver
}

var _ runner.Service = (*GraphQL)(nil)

func (g GraphQL) Start(rn *runner.ServiceRunner) error {
	schema, err := graphql.ParseSchema(rawSchema, &g.gqlResolver,
		graphql.UseFieldResolvers(),
		graphql.UseStringDescriptions())
	if err != nil {
		log.Println(err)
		return err
	}

	relayHandler := relay.Handler{Schema: schema}
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        graphQLPrefix,
			Method:      http.MethodPost,
			HandlerFunc: relayHandler.ServeHTTP,
		},
	})
	return nil
}

func NewGraphQL(gqlResolver gql.Resolver) GraphQL {
	return GraphQL{
		gqlResolver: gqlResolver,
	}
}
