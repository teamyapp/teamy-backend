package entity

type UserLink struct {
	AuthProvider   string
	InternalUserID uint64
	ExternalUserID string
}
