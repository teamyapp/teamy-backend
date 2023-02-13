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

type App struct {
	dataCollector telemetry.DataCollector
	appID         string
	appPrivateKey *rsa.PrivateKey
	jwt           *string
	jwtExpireAt   *time.Time
	installations map[string]*Installation
}

func (a *App) GetInstallation(installationID string) *Installation {
	installation, ok := a.installations[installationID]
	if ok {
		return installation
	}

	installation = newInstallation(a.dataCollector, a, installationID)
	a.installations[installationID] = installation
	return installation
}

func (a *App) getOrRefreshAppJWT(ct context.Context) (string, *errs.Error) {
	if a.jwt != nil && a.jwtExpireAt != nil {
		// JWT token is not expired
		if time.Now().UTC().Before(*a.jwtExpireAt) {
			return *a.jwt, nil
		}
	}

	// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps
	expireAt := time.Now().Add(10 * time.Minute)
	payload := struct {
		IssuedAt int64  `json:"iat"` // unix timestamp (UTC)
		ExpireAt int64  `json:"exp"` // unix timestamp (UTC)
		Issuer   string `json:"iss"` // GitHub App ID
	}{
		// To protect against clock drift between backend and GitHub, set this 1 minute in the past
		IssuedAt: time.Now().Add(-1 * time.Minute).Unix(),
		// no more than 10 minutes in the future based on CURRENT time (independent of IssuedAt)
		ExpireAt: expireAt.Unix(),
		Issuer:   a.appID,
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	payloadMap := make(map[string]interface{})
	err = json.Unmarshal(buf, &payloadMap)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(payloadMap))
	signedStr, err := jwtToken.SignedString(a.appPrivateKey)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	a.jwt = &signedStr
	a.jwtExpireAt = &expireAt
	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("Refreshed app JWT token expiring at %v", a.jwtExpireAt))
	return signedStr, nil
}

func NewApp(dataCollector telemetry.DataCollector, appID string, privateKeyPEM []byte) (*App, *errs.Error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(internalErr)
		return nil, internalErr
	}

	return &App{
		appID:         appID,
		appPrivateKey: privateKey,
		installations: map[string]*Installation{},
	}, nil
}
