package client

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type GithubApp struct {
	dataCollector telemetry.DataCollector
	appID         string
	appPrivateKey *rsa.PrivateKey
	jwt           *string
	jwtExpireAt   *time.Time
	installations map[int]*Installation
}

func (g *GithubApp) GetInstallation(installationID int) *Installation {
	installation, ok := g.installations[installationID]
	if ok {
		return installation
	}

	installation = newInstallation(g.dataCollector, g, installationID)
	g.installations[installationID] = installation
	return installation
}

func (g *GithubApp) getOrRefreshAppJWT(ct context.Context) (string, *errs.Error) {
	if g.jwt != nil && g.jwtExpireAt != nil {
		// JWT token is not expired
		if time.Now().UTC().Before(*g.jwtExpireAt) {
			return *g.jwt, nil
		}
	}

	// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps
	expireAt := time.Now().Add(10 * time.Minute)
	issuedAt := time.Now().Add(-1 * time.Minute)
	payload := struct {
		IssuedAt int64  `json:"iat"` // unix timestamp (UTC)
		ExpireAt int64  `json:"exp"` // unix timestamp (UTC)
		Issuer   string `json:"iss"` // GitHub App ID
	}{
		// To protect against clock drift between backend and GitHub, set this 1 minute in the past
		IssuedAt: issuedAt.Unix(),
		// no more than 10 minutes in the future based on CURRENT time (independent of IssuedAt)
		ExpireAt: expireAt.Unix(),
		Issuer:   g.appID,
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	payloadMap := make(map[string]interface{})
	err = json.Unmarshal(buf, &payloadMap)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(payloadMap))
	signedStr, err := jwtToken.SignedString(g.appPrivateKey)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	g.jwt = &signedStr
	g.jwtExpireAt = &expireAt
	g.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("Refreshed app JWT token expiring at %v", g.jwtExpireAt))
	return signedStr, nil
}

func NewGithubApp(dataCollector telemetry.DataCollector, appID string, privateKeyPEM []byte) (*GithubApp, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(internalErr)
		return nil, internalErr.ToError()
	}

	return &GithubApp{
		appID:         appID,
		appPrivateKey: privateKey,
		installations: map[int]*Installation{},
	}, nil
}
