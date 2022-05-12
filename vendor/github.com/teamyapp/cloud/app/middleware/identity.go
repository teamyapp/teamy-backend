package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/app/ctx"
)

type Identity struct {
	verifyTokenURL string
	handler        http.Handler
}

var _ http.Handler = (*Identity)(nil)

func (i Identity) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	token, err := getBearerToken(request)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(token) > 0 {
		userID, err := i.getUserID(token)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		ct := ctx.NewContextWithUserID(request.Context(), userID)
		request = request.WithContext(ct)
	}

	i.handler.ServeHTTP(writer, request)
}

func (i Identity) getUserID(bearerToken string) (uint64, error) {
	res, err := http.Post(
		i.verifyTokenURL,
		"text/plain",
		bytes.NewReader([]byte(bearerToken)))
	if err != nil {
		return 0, err
	}
	if res.StatusCode == http.StatusUnauthorized {
		return 0, errors.New("invalid access token")
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	userID, err := strconv.ParseUint(string(buf), 10, 64)
	return userID, err
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

func WithIdentity(identityAPIEndpoint string, handler http.Handler) http.Handler {
	verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
	return Identity{
		verifyTokenURL: verifyTokenURL,
		handler:        handler,
	}
}
