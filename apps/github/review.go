package github

type pullRequestReviewAction string

const (
	submittedPullRequestReviewAction       pullRequestReviewAction = "submitted"
	editedRequestedPullRequestReviewAction pullRequestReviewAction = "edited"
	dismissedPullRequestReviewAction       pullRequestReviewAction = "dismissed"
)

type pullRequestReviewState string

const (
	commentedPullRequestReviewState        pullRequestReviewAction = "commented"
	changesRequestedPullRequestReviewState pullRequestReviewAction = "changes_requested"
	approvedPullRequestReviewState         pullRequestReviewAction = "approved"
)

type authorAssociation string

const (
	ownerAuthorAssociation       authorAssociation = "OWNER"
	contributorAuthorAssociation authorAssociation = "CONTRIBUTOR"
	noneAuthorAssociation        authorAssociation = "NONE"
)

type pullRequestReview struct {
	ID                uint64                 `json:"id"`
	NodeID            string                 `json:"node_id"`
	User              user                   `json:"user"`
	Body              string                 `json:"body"`
	CommitID          string                 `json:"commit_id"`
	SubmittedAt       githubTime             `json:"submitted_at"`
	State             pullRequestReviewState `json:"state"`
	AuthorAssociation authorAssociation      `json:"author_association"`
}
