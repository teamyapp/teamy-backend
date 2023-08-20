package entity

type GithubCodeReview struct {
	GithubPullRequestNodeID       string
	GithubReviewerNodeID          string
	InternalCodeReviewTaskID      uint64
	InternalAddressFeedbackTaskID *uint64
	Round                         int
}
