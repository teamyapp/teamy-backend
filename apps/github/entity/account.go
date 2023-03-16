package entity

type accountType string

const (
	UserAccountType         accountType = "User"
	OrganizationAccountType accountType = "Organization"
)

type account struct {
	Login       string      `json:"login"`
	ID          uint64      `json:"id"`
	NodeID      string      `json:"node_id"`
	AvatarURL   string      `json:"avatar_url"`
	GravatarID  string      `json:"gravatar_id"`
	SiteAdmin   bool        `json:"site_admin"`
	Description string      `json:"description"`
	Type        accountType `json:"type"`
}
