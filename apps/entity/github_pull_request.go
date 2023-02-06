package entity

type GithubPullRequest struct {
	NodeID          string
	InternalTaskID  uint64
	Number          int
	URL             string
	RepositoryName  string
	RepositoryOwner string
}
