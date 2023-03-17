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
		getString(g.RepositoryOwner),
		getString(g.RepositoryName),
		getInt(g.Number),
		getString(g.URL),
		getUint64(g.OrganizationID),
	)
}

func getString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	
	return *ptr
}

func getInt(ptr *int) int {
	if ptr == nil {
		return 0
	}
	
	return *ptr
}

func getUint64(ptr *uint64) uint64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
