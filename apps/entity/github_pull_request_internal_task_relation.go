package entity

type GithubPullRequestInternalTaskRelation struct {
	PullRequestNodeID  string
	InternalTaskID     uint64
	InternalTaskLinkID uint64
	AutomaticTracking  bool
}
