package github

type accountType string

const (
	userAccountType         accountType = "User"
	organizationAccountType accountType = "Organization"
)

type account struct {
	Type accountType `json:"type"`
}
