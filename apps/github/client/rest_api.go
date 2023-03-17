package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/apps/github/entity"
)

const (
	gitHubAPIVersion = "2022-11-28"
)

type RESTAPI struct {
	dataCollector telemetry.DataCollector
	httpClient    web.HTTPClient
}

func (r RESTAPI) GetOrganizationByLogin(ct context.Context, installation *Installation, login string) (entity.Organization, *errs.Error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s", login)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		internalErr := errs.NewError(errs.Unknown, err.Error())
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Organization{}, internalErr
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)

	accessToken, internalErr := installation.GetOrRefreshAccessToken(ct)
	if internalErr != nil {
		return entity.Organization{}, internalErr
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := r.httpClient.Do(req)
	if err != nil {
		internalErr = errs.NewError(errs.Unknown, err.Error())
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Organization{}, internalErr
	}

	if res.StatusCode >= errs.HTTPClientErrors && res.StatusCode < errs.HTTPServerErrors {
		if res.StatusCode == http.StatusNotFound {
			internalErr = errs.NewError(errs.NotFound, "Not found")
		} else {
			internalErr = errs.NewError(errs.Unknown, "Unknown")
		}
		return entity.Organization{}, errs.NewError(errs.Unknown, "Unknown")
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = errs.NewError(errs.IO, err.Error())
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Organization{}, internalErr
	}

	var body entity.Organization
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr = errs.NewError(errs.Deserialization, err.Error())
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Organization{}, internalErr
	}

	return body, nil
}

func NewRESTAPI(dataCollector telemetry.DataCollector, httpClient web.HTTPClient) RESTAPI {
	return RESTAPI{
		dataCollector: dataCollector,
		httpClient:    httpClient,
	}
}
