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
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/security"
)

type tokenPayload struct {
	UserID           uint64     `json:"user_id"`
	IssuedAt         *time.Time `json:"issued_at"`
	IsServiceAccount bool       `json:"is_service_account"`
	Secret           *string    `json:"secret"`
}

type Identity struct {
	signInSessionDao  dao.SignInSession
	userLinkDao       dao.UserLink
	serviceAccountDao dao.ServiceAccount
	userIDGenerator   *gen.UniqueNumber
	stateIDGenerator  *gen.UniqueNumber
	jwtAuthority      security.JWTAuthority
	oauthProviders    map[string]oauth.Provider
	accessTokenTLL    time.Duration
}

func (i Identity) VerifyAccessToken(accessToken string) (uint64, bool) {
	payload := tokenPayload{}
	err := i.jwtAuthority.DecodeToken(accessToken, &payload)
	if err != nil {
		return 0, false
	}

	if payload.IsServiceAccount {
		serviceAccount, err := i.serviceAccountDao.FindServiceAccountByID(payload.UserID)
		if err != nil {
			log.Println(err)
			return 0, false
		}

		if serviceAccount.Secret == nil || payload.Secret == nil {
			return 0, false
		}

		return serviceAccount.ID, *serviceAccount.Secret == *payload.Secret
	} else {
		issuedAt := *payload.IssuedAt
		if issuedAt.Add(i.accessTokenTLL).Before(time.Now()) {
			return 0, false
		}
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

	now := time.Now()
	payload := tokenPayload{
		UserID:   userID,
		IssuedAt: &now,
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

func (i Identity) ListServiceAccounts(accountOwnerID uint64) ([]entity.ServiceAccount, error) {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(accountOwnerID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return collect.Map(serviceAccounts, func(serviceAccount entity.ServiceAccount, _ int) entity.ServiceAccount {
		serviceAccount.Secret = nil
		return serviceAccount
	}), nil
}

func (i Identity) CreateServiceAccount(accountOwnerID uint64, serviceAccountName string) error {
	serviceAccountID, err := i.userIDGenerator.GenerateUniqueNumber()
	if err != nil {
		log.Println(err)
		return err
	}

	account := entity.ServiceAccount{
		ID:          serviceAccountID,
		Name:        serviceAccountName,
		OwnerUserID: accountOwnerID,
		CreatedAt:   time.Now().UTC(),
	}

	return i.serviceAccountDao.CreateServiceAccount(account)
}

func (i Identity) GenerateServiceToken(accountOwnerID uint64, serviceAccountID uint64) (string, error) {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(accountOwnerID)
	if err != nil {
		log.Println(err)
		return "", err
	}

	foundServiceAccounts := collect.Filter(serviceAccounts, func(account entity.ServiceAccount) bool {
		return account.ID == serviceAccountID
	})
	if len(foundServiceAccounts) < 1 {
		return "", fmt.Errorf("service account not found: userID=%v, serviceAccountID=%v\n", accountOwnerID, serviceAccountID)
	}

	serviceAccount := serviceAccounts[0]
	secret := randgen.String(randgen.Base62, 10)
	serviceAccount.Secret = &secret
	err = i.serviceAccountDao.UpdateServiceAccount(serviceAccount)
	if err != nil {
		log.Println(err)
		return "", err
	}

	payload := tokenPayload{
		UserID:           serviceAccountID,
		IsServiceAccount: true,
		Secret:           serviceAccount.Secret,
	}

	return i.jwtAuthority.GenerateToken(payload)
}

func (i Identity) DeleteServiceAccount(accountOwnerID uint64, serviceAccountID uint64) error {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(accountOwnerID)
	if err != nil {
		log.Println(err)
		return err
	}

	foundServiceAccounts := collect.Filter(serviceAccounts, func(account entity.ServiceAccount) bool {
		return account.ID == serviceAccountID
	})
	if len(foundServiceAccounts) < 1 {
		return fmt.Errorf("service account not found: userID=%v, serviceAccountID=%v\n", accountOwnerID, serviceAccountID)
	}

	return i.serviceAccountDao.DeleteServiceAccount(serviceAccountID)
}

func NewIdentity(
	signInSessionDao dao.SignInSession,
	userLinkDao dao.UserLink,
	serviceAccountDao dao.ServiceAccount,
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
		signInSessionDao:  signInSessionDao,
		userLinkDao:       userLinkDao,
		serviceAccountDao: serviceAccountDao,
		userIDGenerator:   userIDGenerator,
		stateIDGenerator:  stateIDGenerator,
		jwtAuthority:      jwtAuthority,
		oauthProviders:    oauthProviderMap,
		accessTokenTLL:    accessTokenTLL,
	}, nil
}
