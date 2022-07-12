package github

import (
	"fmt"
)

type pullRequestReviewAction string

const (
	submittedPullRequestReviewAction       pullRequestReviewAction = "submitted"
	editedRequestedPullRequestReviewAction pullRequestReviewAction = "edited"
	dismissedPullRequestReviewAction       pullRequestReviewAction = "dismissed"
)

type pullRequestReviewState string

const (
	commentedPullRequestReviewState        pullRequestReviewState = "commented"
	changesRequestedPullRequestReviewState pullRequestReviewState = "changes_requested"
	approvedPullRequestReviewState         pullRequestReviewState = "approved"
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

func (p pullRequestReview) String() string {
	return fmt.Sprintf(
		`[pullRequestReview
	ID:%v
	NodeID:%v
	User:%v
	Body:%v
	CommitID:%v
	SubmittedAt:%v
	State:%v
	AuthorAssociation:%v]`,
		p.ID,
		p.NodeID,
		p.User,
		p.Body,
		p.CommitID,
		p.SubmittedAt,
		p.State,
		p.AuthorAssociation)
}
