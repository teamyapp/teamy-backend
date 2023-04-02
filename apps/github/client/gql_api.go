package client

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const githubGraphQLAPIEndpoint = "https://api.github.com/graphql"

type GraphQLAPI struct {
	dataCollector telemetry.DataCollector
	graphQLClient *gql.Client
}

type Node[Type any] struct {
	Node Type `json:"node"`
}

// Standard GraphQL error response:
// https://github.com/graphql/graphql-spec/blob/main/spec/Section%207%20--%20Response.md#errors
// Github GraphQL APIs have their own structure
type GithubGraphQLError struct {
	Message    string     `json:"message"`
	Locations  []Location `json:"locations"`
	Extensions Extensions `json:"extensions"`
}

func (e *GithubGraphQLError) internalErr() *errs.Error {
	return errs.NewError(errs.Unknown, fmt.Sprintf("%+v", e))
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (l *Location) String() string {
	return fmt.Sprintf("[Location Line:%d Column:%d]", l.Line, l.Column)
}

type Problem struct {
	Path        []interface{} `json:"path"`
	Explanation string        `json:"explanation"`
}

func (p *Problem) String() string {
	return fmt.Sprintf("[Problem Path:%v Explanation:%s]", p.Path, p.Explanation)
}

type Extensions struct {
	Value    interface{} `json:"value"`
	Problems []Problem   `json:"problems"`
}

func (e *Extensions) String() string {
	return fmt.Sprintf("[Extensions: Value:%v Problems:%v]", e.Value, e.Problems)
}

type UserNode struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

type PullRequestNode struct {
	Number     int            `json:"number"`
	URL        string         `json:"url"`
	Repository RepositoryNode `json:"repository"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	Author     UserNode       `json:"author"`
}

type RepositoryNode struct {
	Name  string              `json:"name"`
	Owner RepositoryOwnerNode `json:"owner"`
}

type RepositoryOwnerNode struct {
	Login string `json:"login"`
}

type UpdatePullRequestInput struct {
	PullRequestID string  `json:"pullRequestId"`
	Body          *string `json:"body,omitempty"`
}

func (g GraphQLAPI) GetPullRequestByNodeID(ct context.Context, installation *Installation, nodeID string) (PullRequestNode, *errs.Error) {
	queryOptions := gql.QueryOptions{
		Query: `
		query getPullRequest($nodeId:ID!) {
			node(id: $nodeId) {
				... on PullRequest {
					number
					body
					repository {
						owner {
							login
						}
						name
					}
					url
					author {
						login
						... on User {
							id
						}
					}
				}

			}
		}`,
		Variables: struct {
			NodeID string `json:"nodeId"`
		}{
			NodeID: nodeID,
		},
	}

	var res gql.GraphQLResponse[Node[PullRequestNode], GithubGraphQLError]
	err := g.query(ct, installation, queryOptions, &res)
	if err != nil {
		return PullRequestNode{}, err
	}

	if len(res.Errors) > 0 {
		internalErrs := collect.Map(res.Errors, func(err GithubGraphQLError, _ int) *errs.Error {
			return err.internalErr()
		})
		mergedInternalErrs := errs.MergeErrs(internalErrs)
		return PullRequestNode{}, mergedInternalErrs
	}

	return res.Data.Node, nil
}

func (g GraphQLAPI) UpdatePullRequest(ct context.Context, installation *Installation, pullRequestInput UpdatePullRequestInput) (PullRequestNode, *errs.Error) {
	mutationOptions := gql.MutationOptions{
		Mutation: `mutation ($input: UpdatePullRequestInput!) {
			updatePullRequest(input: $input) {
				pullRequest{
					body
				}
			}
		  }`,
		Variables: struct {
			Input UpdatePullRequestInput `json:"input"`
		}{
			Input: pullRequestInput,
		},
	}

	type GqlMutationRes struct {
		PullRequest PullRequestNode `json:"pullRequest"`
	}

	var res gql.GraphQLResponse[Node[GqlMutationRes], GithubGraphQLError]
	err := g.mutate(ct, installation, mutationOptions, &res)
	if err != nil {
		return PullRequestNode{}, err
	}

	if len(res.Errors) > 0 {
		internalErrs := collect.Map(res.Errors, func(err GithubGraphQLError, _ int) *errs.Error {
			return err.internalErr()
		})
		mergedInternalErrs := errs.MergeErrs(internalErrs)
		return PullRequestNode{}, mergedInternalErrs
	}

	return res.Data.Node.PullRequest, nil
}

func (g GraphQLAPI) query(
	ct context.Context,
	installation *Installation,
	queryOptions gql.QueryOptions,
	gqlResponse interface{},
) *errs.Error {
	accessToken, err := installation.GetOrRefreshAccessToken(ct)
	if err != nil {
		return err
	}

	headers := withCredential(make(map[string]string), accessToken)
	return g.graphQLClient.Query(ct, githubGraphQLAPIEndpoint, headers, queryOptions, gqlResponse)
}

func (g *GraphQLAPI) mutate(
	ct context.Context,
	installation *Installation,
	mutationOptions gql.MutationOptions,
	gqlResponse interface{},
) *errs.Error {
	accessToken, err := installation.GetOrRefreshAccessToken(ct)
	if err != nil {
		return err
	}

	headers := withCredential(make(map[string]string), accessToken)
	return g.graphQLClient.Mutate(ct, githubGraphQLAPIEndpoint, headers, mutationOptions, gqlResponse)
}

func NewGraphQLAPI(dataCollector telemetry.DataCollector, graphQLClient *gql.Client) GraphQLAPI {
	return GraphQLAPI{
		dataCollector: dataCollector,
		graphQLClient: graphQLClient,
	}
}
