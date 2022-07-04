package github

import (
	"fmt"
)

type commit struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	SHA   string     `json:"sha"`
	User  user       `json:"user"`
	Repo  repository `json:"repo"`
}

func (c commit) String() string {
	return fmt.Sprintf(
		`[commit
	Label:%v
	Ref:%v
	SHA:%v
	User:%v
	Repo:%v]`,
		c.Label,
		c.Ref,
		c.SHA,
		c.User,
		c.Repo)
}
