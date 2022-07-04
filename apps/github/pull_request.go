package github

import (
	"fmt"
)

type pullRequestAction string

const (
	assignedPullRequestAction             pullRequestAction = "assigned"
	autoMergeDisabledPullRequestAction    pullRequestAction = "auto_merge_disabled"
	autoMergeEnabledPullRequestAction     pullRequestAction = "auto_merge_enabled"
	closedPullRequestAction               pullRequestAction = "closed"
	convertedToDraftPullRequestAction     pullRequestAction = "converted_to_draft"
	editedPullRequestAction               pullRequestAction = "edited"
	labeledPullRequestAction              pullRequestAction = "labeled"
	lockedPullRequestAction               pullRequestAction = "locked"
	openedPullRequestAction               pullRequestAction = "opened"
	readyForReviewPullRequestAction       pullRequestAction = "ready_for_review"
	reopenedPullRequestAction             pullRequestAction = "reopened"
	reviewRequestRemovedPullRequestAction pullRequestAction = "review_request_removed"
	reviewRequestedPullRequestAction      pullRequestAction = "review_requested"
	synchronizePullRequestAction          pullRequestAction = "synchronize"
	unassignedPullRequestAction           pullRequestAction = "unassigned"
	unlabeledPullRequestAction            pullRequestAction = "unlabeled"
	unlockedPullRequestAction             pullRequestAction = "unlocked"
)

type pullRequestState string

const (
	openPullRequestState   pullRequestState = "open"
	closedPullRequestState                  = "closed"
)

type pullRequest struct {
	ID                 int              `json:"id"`
	NodeID             string           `json:"node_id"`
	Number             int              `json:"number"`
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
	Draft              bool             `json:"draft"`
	Head               commit           `json:"head"`
	Base               commit           `json:"base"`
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
