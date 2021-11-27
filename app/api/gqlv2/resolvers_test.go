package gqlv2_test

import (
	"log"
	"net/http"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2"
)

func Test(t *testing.T) {
	handler := &relay.Handler{
		Schema: graphql.MustParseSchema(
			gqlv2.RawSchema(), &gqlv2.Root{
				Deps: gqlv2.Dependencies{
					Data: gqlv2.Read("./data.json"),
				},
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
			writer.Write(gqlv2.QraphiQL())
			return
		}

		handlerFunc.ServeHTTP(writer, request)
	}
}
