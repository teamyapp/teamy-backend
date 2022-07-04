package api

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/core/api/gql"
)

//go:embed gql/schema.graphql
var rawSchema string

type GraphQL struct {
	identityAPIEndpoint string
	resolver            gql.Resolver
}

var _ runner.Service = (*GraphQL)(nil)

func (g GraphQL) Start(runner *runner.ServiceRunner) error {
	schema, err := graphql.ParseSchema(rawSchema, &g.resolver,
		graphql.UseFieldResolvers(),
		graphql.UseStringDescriptions())
	if err != nil {
		log.Println(err)
		return err
	}

	relayHandler := relay.Handler{Schema: schema}
	runner.RegisterWebRoutes([]web.Route{
		{
			Path:        graphQLPrefix,
			Method:      http.MethodPost,
			HandlerFunc: middleware.WithWebIdentity(g.identityAPIEndpoint, relayHandler.ServeHTTP).ServeHTTP,
		},
	})

	return nil
}

func NewGraphQL(identityAPIEndpoint string, resolver gql.Resolver) GraphQL {
	return GraphQL{
		identityAPIEndpoint: identityAPIEndpoint,
		resolver:            resolver,
	}
}
