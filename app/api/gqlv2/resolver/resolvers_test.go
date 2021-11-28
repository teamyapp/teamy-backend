package resolver_test

import (
	"log"
	"net/http"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
)

func Test(t *testing.T) {
	deps := &resolver.Dependencies{
		Data: resolver.Read("./data.json"),
	}
	handler := &relay.Handler{
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
