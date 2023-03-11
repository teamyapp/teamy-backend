package entity

import "fmt"

type GithubPullRequest struct {
	NodeID          string
	InternalTaskID  uint64
	RepositoryOwner *string
	RepositoryName  *string
	Number          *int
	URL             *string
	OrganizationID  *uint64
}

func (g *GithubPullRequest) String() string {
	return fmt.Sprintf(
		"[GithubPullRequest NodeID:%v InternalTaskID:%v RepositoryOwner:%v RepositoryName:%v Number:%v URL:%v OrganizationID:%v]",
		g.NodeID,
		g.InternalTaskID,
		g.RepositoryOwner,
		g.RepositoryName,
		g.Number,
		g.URL,
		g.OrganizationID,
	)
}
