package entity

type SignInSessionType string

const (
	UnknownUserSignInSessionType SignInSessionType = "UNKNOWN_USER_SIGN_IN"
	LinkUsersSignInSessionType   SignInSessionType = "LINK_USERS_SIGN_IN"
)

type SignInSession struct {
	ID             uint64
	Type           SignInSessionType
	InternalUserID *uint64
	RedirectURL    string
}
