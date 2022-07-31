package github

import (
	"fmt"
)

type visibility string

const (
	publicVisibility  visibility = "public"
	privateVisibility visibility = "private"
)

type repository struct {
	ID                       uint64      `json:"id"`
	NodeID                   string      `json:"node_id"`
	Name                     string      `json:"name"`
	FullName                 string      `json:"full_name"`
	Private                  bool        `json:"private"`
	Owner                    user        `json:"owner"`
	Description              string      `json:"description"`
	Fork                     bool        `json:"fork"`
	CreatedAt                githubTime  `json:"created_at"`
	UpdatedAt                *githubTime `json:"updated_at"`
	PushedAt                 *githubTime `json:"pushed_at"`
	Homepage                 *string     `json:"homepage"`
	Size                     int         `json:"size"`
	StargazersCount          int         `json:"stargazers_count"`
	WatchersCount            int         `json:"watchers_count"`
	Language                 *string     `json:"language"`
	HasIssues                bool        `json:"has_issues"`
	HasProjects              bool        `json:"has_projects"`
	HasDownloads             bool        `json:"has_downloads"`
	HasWiki                  bool        `json:"has_wiki"`
	HasPages                 bool        `json:"has_pages"`
	ForksCount               int         `json:"forks_count"`
	Archived                 bool        `json:"archived"`
	Disabled                 bool        `json:"disabled"`
	OpenIssuesCount          int         `json:"open_issues_count"`
	License                  *string     `json:"license"`
	AllowForking             bool        `json:"allow_forking"`
	IsTemplate               bool        `json:"is_template"`
	WebCommitSignOffRequired bool        `json:"web_commit_signoff_required"`
	Topics                   []string    `json:"topics"`
	Visibility               visibility
	Forks                    int    `json:"forks"`
	OpenIssues               int    `json:"open_issues"`
	Watchers                 int    `json:"watchers"`
	DefaultBranch            string `json:"default_branch"`
}

func (r repository) String() string {
	return fmt.Sprintf(
		`[repository
	ID:%v
	NodeID:%v
	Name:%v
	FullName:%v
	Private:%v
	Owner:%v
	Description:%v
	Fork:%v
	CreatedAt:%v
	UpdatedAt:%v
	PushedAt:%v
	Homepage:%v
	Size:%v
	StargazersCount:%v
	WatchersCount:%v
	Language:%v
	HasIssues:%v
	HasProjects:%v
	HasDownloads:%v
	HasWiki:%v
	HasPages:%v
	ForksCount:%v
	Archived:%v
	Disabled:%v
	OpenIssuesCount:%v
	License:%v
	AllowForking:%v
	IsTemplate:%v
	WebCommitSignOffRequired:%v
	Topics:%v
	Visibility:%v
	Forks:%v
	OpenIssues:%v
	Watchers:%v
	DefaultBranch:%v]`,
		r.ID,
		r.NodeID,
		r.Name,
		r.FullName,
		r.Private,
		r.Owner,
		r.Description,
		r.Fork,
		r.CreatedAt,
		r.UpdatedAt,
		r.PushedAt,
		r.Homepage,
		r.Size,
		r.StargazersCount,
		r.WatchersCount,
		r.Language,
		r.HasIssues,
		r.HasProjects,
		r.HasDownloads,
		r.HasWiki,
		r.HasPages,
		r.ForksCount,
		r.Archived,
		r.Disabled,
		r.OpenIssuesCount,
		r.License,
		r.AllowForking,
		r.IsTemplate,
		r.WebCommitSignOffRequired,
		r.Topics,
		r.Visibility,
		r.Forks,
		r.OpenIssues,
		r.Watchers,
		r.DefaultBranch)
}
