package gql

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
)

//go:embed schema.graphqls
var rawSchema string

func NewServer(port int) (http.Server, error) {
	res := resolver.Resolver{}
	schema, err := graphql.ParseSchema(rawSchema, &res)
	if err != nil {
		return http.Server{}, err
	}

	relayHandler := relay.Handler{Schema: schema}
	mux := http.ServeMux{}
	mux.Handle("/query", &relayHandler)
	addr := fmt.Sprintf(":%d", port)
	return http.Server{Addr: addr, Handler: &mux}, nil
}
