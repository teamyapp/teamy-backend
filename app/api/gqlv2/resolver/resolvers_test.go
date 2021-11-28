package resolver_test

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
)

func Test(t *testing.T) {
	deps := resolver.Dependencies{
		Data: resolver.Read("./data.json"),
	}
	handler := &Handler{
		Schema: graphql.MustParseSchema(
			gqlv2.RawSchema(), &resolver.Root{
				Deps: deps,
			},
			graphql.UseFieldResolvers(),
			graphql.UseStringDescriptions(),
		),
	}
	http.Handle("/", includeGraphiQLIDE(handler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func includeGraphiQLIDE(handlerFunc http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writer.WriteHeader(http.StatusOK)
			writer.Write(resolver.QraphiQL())
			return
		}

		handlerFunc.ServeHTTP(writer, request)
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

	response := h.Schema.Exec(r.Context(), params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
