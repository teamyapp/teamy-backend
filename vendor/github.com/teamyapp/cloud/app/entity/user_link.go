package entity

type UserLink struct {
	AuthProvider      string `json:"authProvider"`
	InternalUserID    uint64 `json:"internalUserId"`
	ExternalUserID    string `json:"externalUserId"`
	ExternalUserLabel string `json:"externalUserLabel"`
}
