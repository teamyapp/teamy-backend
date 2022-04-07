package gqlv2

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
	"github.com/teamyapp/teamy-backend/app/log"
)

//go:embed schema.graphql
var rawSchema string

//go:embed GraphQLIDE.html
var graphIDEHTML []byte

func NewServer(identityAPIEndpoint string, res resolver.Resolver, port int) (http.Server, error) {
	schema, err := graphql.ParseSchema(rawSchema, &res,
		graphql.UseFieldResolvers(),
		graphql.UseStringDescriptions())
	if err != nil {
		return http.Server{}, err
	}

	handler := identity.WithMiddleware(identityAPIEndpoint, &relay.Handler{Schema: schema})
	mux := http.ServeMux{}
	mux.HandleFunc("/graphql", requestID(enableCORS(handler.ServeHTTP)))
	addr := fmt.Sprintf(":%d", port)
	return http.Server{Addr: addr, Handler: &mux}, nil
}

func enableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
	// TODO: move into [One] to encourage reuse
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if request.Method == http.MethodOptions {
			return
		}

		handlerFunc(writer, request)
	}
}

const requestIDKey = "request-id"

func requestID(handlerFunc http.HandlerFunc) http.HandlerFunc {
	// TODO: move into [One] to encourage reuse
	return func(writer http.ResponseWriter, request *http.Request) {
		reqID, err := uuid.NewRandom()
		if err != nil {
			log.Info(err)
			handlerFunc(writer, request)
		} else {
			reqIDStr := reqID.String()
			ctx := request.Context()
			ctx = context.WithValue(ctx, requestIDKey, reqIDStr)
			log.Info(ctx, "new request")
			newRequest := request.WithContext(ctx)
			handlerFunc(writer, newRequest)
		}
	}
}
