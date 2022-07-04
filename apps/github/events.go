package github

import (
	"fmt"
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
