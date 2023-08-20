package entity

import (
	"fmt"
)

type pullRequestAction string

const (
	AssignedPullRequestAction             pullRequestAction = "assigned"
	AutoMergeDisabledPullRequestAction    pullRequestAction = "auto_merge_disabled"
	AutoMergeEnabledPullRequestAction     pullRequestAction = "auto_merge_enabled"
	ClosedPullRequestAction               pullRequestAction = "closed"
	ConvertedToDraftPullRequestAction     pullRequestAction = "converted_to_draft"
	EditedPullRequestAction               pullRequestAction = "edited"
	LabeledPullRequestAction              pullRequestAction = "labeled"
	LockedPullRequestAction               pullRequestAction = "locked"
	OpenedPullRequestAction               pullRequestAction = "opened"
	ReadyForReviewPullRequestAction       pullRequestAction = "ready_for_review"
	ReopenedPullRequestAction             pullRequestAction = "reopened"
	ReviewRequestRemovedPullRequestAction pullRequestAction = "review_request_removed"
	ReviewRequestedPullRequestAction      pullRequestAction = "review_requested"
	SynchronizePullRequestAction          pullRequestAction = "synchronize"
	UnassignedPullRequestAction           pullRequestAction = "unassigned"
	UnlabeledPullRequestAction            pullRequestAction = "unlabeled"
	UnlockedPullRequestAction             pullRequestAction = "unlocked"
)

type pullRequestState string

const (
	OpenPullRequestState   pullRequestState = "open"
	ClosedPullRequestState pullRequestState = "closed"
)

type pullRequest struct {
	ID                 int              `json:"id"`
	NodeID             string           `json:"node_id"`
	Number             int              `json:"number"`
	HtmlURL            string           `json:"html_url"`
	URL                string           `json:"url"`
	State              pullRequestState `json:"state"`
	Locked             bool             `json:"locked"`
	Title              string           `json:"title"`
	User               user             `json:"user"`
	Body               string           `json:"body"`
	CreatedAt          githubTime       `json:"created_at"`
	UpdatedAt          *githubTime      `json:"updated_at"`
	ClosedAt           *githubTime      `json:"closed_at"`
	MergedAt           *githubTime      `json:"merged_at"`
	MergeCommitSHA     *string          `json:"merge_commit_sha"`
	Assignee           *user            `json:"assignee"`
	Assignees          []user           `json:"assignees"`
	RequestedReviewers []user           `json:"requested_reviewers"`
	Labels             []string         `json:"labels"`
	Head               commit           `json:"head"`
	Base               commit           `json:"base"`
	Draft              bool             `json:"draft"`
	Merged             bool             `json:"merged"`
	Mergeable          *bool            `json:"mergeable"`
	Rebaseable         *bool            `json:"rebaseable"`
}

func (p pullRequest) String() string {
	return fmt.Sprintf(
		`[pullRequest
	ID:%v
	NodeID:%v
	Number:%v
	State:%v
	Locked:%v
	Title:%v
	User:%v
	Body:%v
	CreatedAt:%v
	UpdatedAt:%v
	ClosedAt:%v
	MergedAt:%v
	MergeCommitSHA:%v,
	Assignee:%v
	Assignees:%v
	RequestedReviewers:%v
	Labels:%v
	Draft:%v
	Head:%v
	Base:%v]`,
		p.ID,
		p.NodeID,
		p.Number,
		p.State,
		p.Locked,
		p.Title,
		p.User,
		p.Body,
		p.CreatedAt,
		p.UpdatedAt,
		p.ClosedAt,
		p.MergedAt,
		p.MergeCommitSHA,
		p.Assignee,
		p.Assignees,
		p.RequestedReviewers,
		p.Labels,
		p.Draft,
		p.Head,
		p.Base)
}
