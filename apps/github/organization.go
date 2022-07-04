package github

import (
	"fmt"
)

type organization struct {
	Login       string `json:"login"`
	ID          uint64 `json:"id"`
	NodeID      string `json:"node_id"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
}

func (o organization) String() string {
	return fmt.Sprintf(
		`[organization
	Login:%v
	ID:%v
	NodeID:%v
	AvatarURL:%v
	Description:%v]`,
		o.Login,
		o.ID,
		o.NodeID,
		o.AvatarURL,
		o.Description)
}
