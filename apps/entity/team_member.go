package entity

type TeamMember struct {
	UserID           uint64  `json:"userId"`
	FirstName        string  `json:"firstName"`
	LastName         string  `json:"lastName"`
	ProfileURL       *string `json:"profileUrl"`
	HasGithubAccount bool    `json:"hasGithubAccount"`
}
