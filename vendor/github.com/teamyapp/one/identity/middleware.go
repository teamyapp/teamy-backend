package identity

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/one/entity"
)

type Middleware struct {
	verifyTokenURL string
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
	res, err := http.Post(
		"http://localhost:9500/identity/verify-token",
		"text/plain", bytes.NewReader([]byte(bearerToken)))
	if err != nil {
		return 0, err
	}
	if res.StatusCode == http.StatusUnauthorized {
		return -1, errors.New("invalid access token")
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	userID, err := strconv.Atoi(string(buf))
	return entity.ID(userID), err
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

func WithMiddleware(identityAPIEndpoint string, handler http.Handler) Middleware {
	verifyTokenURL := fmt.Sprintf("%s/identity/verify-token", identityAPIEndpoint)
	return Middleware{
		verifyTokenURL: verifyTokenURL,
		handler: handler,
	}
}
