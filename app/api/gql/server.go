package gql

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
	"github.com/teamyapp/teamy-backend/app/service"
)

//go:embed schema.graphqls
var rawSchema string

func NewServer(executionService service.Execution, port int) (http.Server, error) {
	res := resolver.NewResolver(executionService)
	schema, err := graphql.ParseSchema(rawSchema, &res)
	if err != nil {
		log.Println(err)
		return http.Server{}, err
	}

	handler := identity.WithMiddleware(&relay.Handler{Schema: schema})
	mux := http.ServeMux{}
	mux.Handle("/graphql", handler)
	addr := fmt.Sprintf(":%d", port)
	return http.Server{Addr: addr, Handler: &mux}, nil
}
