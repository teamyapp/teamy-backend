package gql

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/api/gql/operations"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
	"github.com/teamyapp/teamy-backend/app/log"
)

//go:embed schema.graphqls
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

	handler := identity.WithMiddleware(identityAPIEndpoint, &Handler{Schema: schema})
	mux := http.ServeMux{}
	mux.HandleFunc("/graphql", requestID(enableCORS(includeGraphiQLIDE(handler.ServeHTTP))))
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

func includeGraphiQLIDE(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write(graphIDEHTML)
			return
		}

		handlerFunc(writer, request)
	}
}

type Handler struct {
	Schema *graphql.Schema
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// handle server stored queries
	// Execute custom queries if specified
	if params.Query == "" {
		params.Query = operations.Operations()
	}
	ctx := r.Context()
	// generate request id

	log.Info(ctx, "begin GraphQL")
	response := h.Schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	if response.Extensions == nil {
		response.Extensions = make(map[string]interface{})
	}
	response.Extensions["request-id"] = ctx.Value(requestIDKey)
	log.Info(ctx, "end GraphQL")
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
