package identity

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/one/entity"
)

type Middleware struct {
	handler http.Handler
}

var _ http.Handler = (*Middleware)(nil)

func (w Middleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	token, err := getBearerToken(request)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(token) > 0 {
		userID, err := getUserID(token)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := newContext(request.Context(), userID)
		request = request.WithContext(ctx)
	}

	w.handler.ServeHTTP(writer, request)
}

func getUserID(bearerToken string) (entity.ID, error) {
	// TODO: invoke Identity service API
	num, err := strconv.Atoi(bearerToken)
	return entity.ID(num), err
}

func getBearerToken(request *http.Request) (string, error) {
	value := request.Header.Get("Authorization")
	if len(value) == 0 {
		return "", nil
	}

	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid Authorization header format: %s\n", value)
		log.Println(err)
		return "", err
	}

	if parts[0] != "Bearer" {
		err := fmt.Errorf("invalid Authorization header format: %v\n", parts)
		log.Println(err)
		return "", err
	}
	return parts[1], nil
}

func WithMiddleware(handler http.Handler) Middleware {
	return Middleware{handler: handler}
}
