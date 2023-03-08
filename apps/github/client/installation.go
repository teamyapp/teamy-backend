package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const (
	gitHubAPIVersion = "2022-11-28"
)

type Installation struct {
	dataCollector telemetry.DataCollector
	app           *GithubApp
	id            int
	accessToken   *string
	expireAt      *time.Time
}

func (i *Installation) GetOrRefreshAccessToken(ct context.Context) (string, *errs.Error) {
	if i.accessToken != nil && i.expireAt != nil {
		// access token is not expired
		if time.Now().UTC().Before(*i.expireAt) {
			return *i.accessToken, nil
		}
	}

	appJWT, internalErr := i.app.getOrRefreshAppJWT(ct)
	if internalErr != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	githubAppAccessTokenURL := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", i.id)
	req, err := http.NewRequest("POST", githubAppAccessTokenURL, nil)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)

	bearerToken := fmt.Sprintf("Bearer %s", appJWT)
	req.Header.Set("Authorization", bearerToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return "", &errs.Error{
			Code: errs.Unauthenticated,
		}
	case http.StatusForbidden:
		return "", &errs.Error{
			Code: errs.PermissionDenied,
		}
	case http.StatusNotFound:
		return "", &errs.Error{
			Code: errs.NotFound,
		}
	case http.StatusUnprocessableEntity:
		return "", &errs.Error{
			Code: errs.Unknown,
		}
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	// GitHub's authentication token format:
	// ghp_ for Personal Access Tokens
	// gho_ for OAuth Access tokens
	// ghu_ for GitHub App user-to-server tokens
	// ghs_ for GitHub App server-to-server tokens
	// ghr_ for GitHub App refresh tokens
	var body struct {
		Token       string    `json:"token"` // ghs_
		ExpiresAt   time.Time `json:"expires_at"`
		Permissions struct {
			Contents     string `json:"contents"`
			Metadata     string `json:"metadata"`
			PullRequests string `json:"pull_requests"`
			Statuses     string `json:"statuses"`
		} `json:"permissions"`
		RepositorySelection string `json:"repository_selection"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	i.accessToken = &body.Token
	i.expireAt = &body.ExpiresAt
	i.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("Refreshed access token expiring at %v", i.expireAt))
	return body.Token, nil
}

func newInstallation(dataCollector telemetry.DataCollector, app *GithubApp, id int) *Installation {
	return &Installation{
		dataCollector: dataCollector,
		app:           app,
		id:            id,
	}
}
