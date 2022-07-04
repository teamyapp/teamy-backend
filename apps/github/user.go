package github

import (
	"fmt"
)

type user struct {
	Login      string `json:"login"`
	ID         uint64 `json:"id"`
	NodeID     string `json:"node_id"`
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	SiteAdmin  bool   `json:"site_admin"`
}

func (u user) String() string {
	return fmt.Sprintf(
		`[user
	Login:%v
	ID:%v
	NodeID:%v
	AvatarURL:%v
	GravatarID:%v
	SiteAdmin:%v]`,
		u.Login,
		u.ID,
		u.NodeID,
		u.AvatarURL,
		u.GravatarID,
		u.SiteAdmin)
}
