package entity

import (
	"fmt"
)

type EventType string

const (
	PullRequestEventType         EventType = "pull_request"
	PullRequestReviewEventType   EventType = "pull_request_review"
	PullRequestStatusesEventType EventType = "statuses"
)

type Event struct {
	Sender       account      `json:"sender"`
	Repository   repository   `json:"repository"`
	Organization organization `json:"organization"`
	Installation installation `json:"installation"`
}

type PullRequestEvent struct {
	Action      pullRequestAction `json:"action"`
	Number      int               `json:"number"`
	PullRequest pullRequest       `json:"pull_request"`
}

func (p PullRequestEvent) String() string {
	return fmt.Sprintf(
		`[PullRequestEvent
	Action:%v
	Number:%v
	PullRequest:%v]`,
		p.Action,
		p.Number,
		p.PullRequest)
}

type PullRequestReviewEvent struct {
	Action      pullRequestReviewAction `json:"action"`
	Review      pullRequestReview       `json:"review"`
	PullRequest pullRequest             `json:"pull_request"`
}

func (p PullRequestReviewEvent) String() string {
	return fmt.Sprintf(
		`[PullRequestReviewEvent
	Action:%v
	Review:%v
	PullRequest:%v]`,
		p.Action,
		p.Review,
		p.PullRequest)
}
