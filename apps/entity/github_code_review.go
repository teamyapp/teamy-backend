package entity

type GithubCodeReview struct {
	GithubPullRequestNodeID       string
	GithubReviewerID              uint64
	InternalCodeReviewTaskID      uint64
	InternalAddressFeedbackTaskID *uint64
	Round                         int
}
