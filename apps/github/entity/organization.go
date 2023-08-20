package entity

import (
	"fmt"
)

type Organization struct {
	Login       string `json:"login"`
	ID          uint64 `json:"id"`
	NodeID      string `json:"node_id"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
}

func (o Organization) String() string {
	return fmt.Sprintf(
		`[Organization
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
