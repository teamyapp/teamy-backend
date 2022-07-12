package github

import (
	"fmt"
)

type eventType string

const (
	pullRequestEventType         eventType = "pull_request"
	pullRequestReviewEventType   eventType = "pull_request_review"
	pullRequestStatusesEventType eventType = "statuses"
)

type event struct {
	Sender       account      `json:"sender"`
	Repository   repository   `json:"repository"`
	Organization organization `json:"organization"`
	Installation installation `json:"installation"`
}

type pullRequestEvent struct {
	Action      pullRequestAction `json:"action"`
	Number      int               `json:"number"`
	PullRequest pullRequest       `json:"pull_request"`
}

func (p pullRequestEvent) String() string {
	return fmt.Sprintf(
		`[pullRequestEvent
	Action:%v
	Number:%v
	PullRequest:%v]`,
		p.Action,
		p.Number,
		p.PullRequest)
}

type pullRequestReviewEvent struct {
	Action      pullRequestReviewAction `json:"action"`
	Review      pullRequestReview       `json:"review"`
	PullRequest pullRequest             `json:"pull_request"`
}

func (p pullRequestReviewEvent) String() string {
	return fmt.Sprintf(
		`[pullRequestReviewEvent
	Action:%v
	Review:%v
	PullRequest:%v]`,
		p.Action,
		p.Review,
		p.PullRequest)
}
