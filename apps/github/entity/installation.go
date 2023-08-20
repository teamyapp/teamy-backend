package entity

import (
	"fmt"
)

type installation struct {
	ID      uint64  `json:"id"`
	NodeID  string  `json:"node_id"`
	Account account `json:"account"`
}

func (i installation) String() string {
	return fmt.Sprintf(
		`[installation
	ID:%v
	NodeID:%v
	Account:%v]`,
		i.ID,
		i.NodeID,
		i.Account)
}
