package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
)

const GitHubName = "github"

// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#web-application-flow
var githubAuthorizationURL = "https://github.com/login/oauth/authorize"
var githubAccessTokenURL = "https://github.com/login/oauth/access_token"
var githubUserURL = "https://api.github.com/user"

type GitHub struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func (g GitHub) GetName() string {
	return GitHubName
}

func (g GitHub) GetUser(authorizationCode string) (entity.ExternalUser, error) {
	// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-
	// oauth-apps#2-users-are-redirected-back-to-your-site-by-github
	accessToken, err := g.getAccessToken(authorizationCode)
	if err != nil {
		return entity.ExternalUser{}, err
	}

	// https://docs.github.com/en/rest/users/emails#list-email-addresses-for-the-authenticated-user
	req, err := http.NewRequest("GET", githubUserURL, nil)
	if err != nil {
		return entity.ExternalUser{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return entity.ExternalUser{}, err
	}

	if res.StatusCode > 300 || res.StatusCode < 200 {
		return entity.ExternalUser{}, fmt.Errorf("fail to obtain %s user ID: HTTPStatusCode=%v", g.GetName(), res.StatusCode)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return entity.ExternalUser{}, err
	}

	var body struct {
		UserID uint64 `json:"id"`
		Login  string `json:"login"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		return entity.ExternalUser{}, err
	}

	return entity.ExternalUser{
		ID:    strconv.FormatUint(body.UserID, 10),
		Label: body.Login,
	}, err
}

func (g GitHub) GetStateID(request *http.Request) (uint64, error) {
	return strconv.ParseUint(request.URL.Query().Get("state"), 10, 64)
}

func (g GitHub) GetAuthorizationCode(request *http.Request) string {
	return request.URL.Query().Get("code")
}

func (g GitHub) GetSignInURL(stateID uint64) (string, error) {
	baseURL, err := url.Parse(githubAuthorizationURL)
	if err != nil {
		return "", err
	}

	query := baseURL.Query()
	query.Add("client_id", g.clientID)
	query.Add("redirect_uri", g.redirectURI)
	query.Add("scope", "read:user")
	query.Add("state", strconv.Itoa(int(stateID)))
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func (g GitHub) getAccessToken(authorizationCode string) (string, error) {
	tokenBody := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}{
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		Code:         authorizationCode,
	}

	buf, err := json.Marshal(tokenBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", githubAccessTokenURL, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode > 300 || res.StatusCode < 200 {
		return "", fmt.Errorf("fail to obtain %s access token: HTTPStatusCode=%v", g.GetName(), res.StatusCode)
	}

	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	body := struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(buf, &body)
	return body.AccessToken, err
}

func NewGitHub(webAPIBaseURL string, clientID string, clientSecret string) GitHub {
	return GitHub{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, GitHubName),
	}
}
