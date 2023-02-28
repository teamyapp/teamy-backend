package entity

type GithubPullRequest struct {
	NodeID          string
	InternalTaskID  uint64
	RepositoryOwner string
	RepositoryName  string
	Number          int
	URL             string
	OrganizationID  uint64
}
