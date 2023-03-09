package entity

import "fmt"

type GithubPullRequest struct {
	NodeID          string
	InternalTaskID  uint64
	RepositoryOwner string
	RepositoryName  string
	Number          int
	URL             string
	OrganizationID  uint64
}

func (g *GithubPullRequest) String() string {
	return fmt.Sprintf(
		"[GithubPullRequest NodeID:%s InternalTaskID:%d RepositoryOwner:%s RepositoryName:%s Number:%d URL:%s OrganizationID:%s]",
		g.NodeID,
		g.InternalTaskID,
		g.RepositoryOwner,
		g.RepositoryName,
		g.Number,
		g.URL,
		g.OrganizationID,
	)
}
