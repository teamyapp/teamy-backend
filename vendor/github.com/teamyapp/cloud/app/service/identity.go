package service

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/security"
)

type tokenPayload struct {
	UserID   uint64 `json:"user_id"`
	IssuedAt string `json:"issued_at"`
}

type Identity struct {
	signInSessionDao dao.SignInSession
	userLinkDao      dao.UserLink
	userIDGenerator  *gen.UniqueNumber
	stateIDGenerator *gen.UniqueNumber
	jwtAuthority     security.JWTAuthority
	oauthProviders   map[string]oauth.Provider
	accessTokenTLL   time.Duration
}

func (i Identity) VerifyAccessToken(accessToken string) (uint64, bool) {
	payload := tokenPayload{}
	err := i.jwtAuthority.DecodeToken(accessToken, &payload)
	if err != nil {
		return 0, false
	}

	tm, err := time.Parse(time.RFC3339, payload.IssuedAt)
	if err != nil {
		return 0, false
	}

	if tm.Add(i.accessTokenTLL).Before(time.Now()) {
		return 0, false
	}

	return payload.UserID, true
}

func (i Identity) GenerateSignInURL(providerName string, redirectURL string) (string, error) {
	provider, err := i.GetOAuthProvider(providerName)
	if err != nil {
		return "", err
	}

	sessionID, err := i.stateIDGenerator.GenerateUniqueNumber()
	if err != nil {
		return "", err
	}

	session := entity.SignInSession{
		ID:          sessionID,
		RedirectURL: redirectURL,
	}

	err = i.signInSessionDao.Add(session)
	if err != nil {
		return "", err
	}

	signInURL, err := provider.GetSignInURL(sessionID)
	if err == nil {
		log.Printf("sign in URL: %s", signInURL)
	}

	return signInURL, err
}

func (i Identity) GetOAuthProvider(providerName string) (oauth.Provider, error) {
	provider, ok := i.oauthProviders[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", provider)
	}

	return provider, nil
}

func (i Identity) FinishOAuthSignIn(providerName string, authorizationCode string, sessionID uint64) (string, error) {
	provider, err := i.GetOAuthProvider(providerName)
	if err != nil {
		return "", err
	}

	externalUser, err := provider.GetUser(authorizationCode)
	if err != nil {
		return "", err
	}

	userID, err := i.getOrLinkInternalUserID(providerName, externalUser.ID)
	if err != nil {
		return "", err
	}

	payload := tokenPayload{
		UserID:   userID,
		IssuedAt: time.Now().Format(time.RFC3339),
	}

	accessToken, err := i.jwtAuthority.GenerateToken(payload)
	if err != nil {
		return "", err
	}

	session, err := i.signInSessionDao.FindByID(sessionID)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(session.RedirectURL)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Add("accessToken", accessToken)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (i Identity) getOrLinkInternalUserID(authProvider string, externalUserID string) (uint64, error) {
	userLink, err := i.userLinkDao.FindByExternalUserID(authProvider, externalUserID)
	switch err.(type) {
	case nil:
		return userLink.InternalUserID, nil
	case dao.ErrNotFound:
		internalUserID, err := i.userIDGenerator.GenerateUniqueNumber()
		if err != nil {
			return 0, err
		}

		userLink = entity.UserLink{
			AuthProvider:   authProvider,
			InternalUserID: internalUserID,
			ExternalUserID: externalUserID,
		}
		return internalUserID, i.userLinkDao.Add(userLink)
	default:
		return 0, err
	}
}

func NewIdentity(
	signInSessionDao dao.SignInSession,
	userLinkDao dao.UserLink,
	uniqueNumberFactory gen.UniqueNumberFactory,
	jwtAuthority security.JWTAuthority,
	oauthProviders []oauth.Provider,
	accessTokenTLL time.Duration,
) (Identity, error) {
	userIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("userID")
	if err != nil {
		return Identity{}, err
	}

	stateIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("stateID")
	if err != nil {
		return Identity{}, err
	}

	oauthProviderMap := make(map[string]oauth.Provider)
	for _, oauthProvider := range oauthProviders {
		oauthProviderMap[oauthProvider.GetName()] = oauthProvider
	}

	return Identity{
		signInSessionDao: signInSessionDao,
		userLinkDao:      userLinkDao,
		userIDGenerator:  userIDGenerator,
		stateIDGenerator: stateIDGenerator,
		jwtAuthority:     jwtAuthority,
		oauthProviders:   oauthProviderMap,
		accessTokenTLL:   accessTokenTLL,
	}, nil
}
