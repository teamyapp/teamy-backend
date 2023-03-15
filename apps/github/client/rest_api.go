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
)

const (
	gitHubAPIVersion = "2022-11-28"
)

type RestAPI struct {
	dataCollector telemetry.DataCollector
	httpClient    web.HTTPClient
}

func (r RestAPI) GetOrgIDByOrgName(ct context.Context, installation *Installation, orgName string) (uint64, *errs.Error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s", orgName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.Unknown,
			Message: err.Error(),
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)

	accessToken, internalErr := installation.GetOrRefreshAccessToken(ct)
	if internalErr != nil {
		return 0, internalErr
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := r.httpClient.Do(req)
	if err != nil {
		internalErr = &errs.Error{
			Code:    errs.Unknown,
			Message: err.Error(),
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	if res.StatusCode == http.StatusNotFound {
		return 0, &errs.Error{
			Code: errs.NotFound,
		}
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:    errs.IO,
			Message: err.Error(),
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	var body struct {
		ID uint64 `json:"id"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr = &errs.Error{
			Code:    errs.IO,
			Message: err.Error(),
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	return body.ID, nil
}

func NewRestAPI(dataCollector telemetry.DataCollector, httpClient web.HTTPClient) RestAPI {
	return RestAPI{
		dataCollector: dataCollector,
		httpClient:    httpClient,
	}
}
