package client

import (
	"context"
	"errors"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const githubGraphQLAPIEndpoint = "https://api.github.com/graphql"

type GraphQLAPI struct {
	dataCollector telemetry.DataCollector
	graphQLClient gql.Client
}

type Node[Type any] struct {
	Node Type `json:"node"`
}

// Error is GitHub GraphQL API error format:
// https://github.com/graphql/graphql-spec/blob/main/spec/Section%207%20--%20Response.md#errors
// where entry "extensions" is GitHub GraphQL API specific error extension attributes
type Error struct {
	Message    string     `json:"message"`
	Locations  []Location `json:"locations"`
	Extensions Extensions `json:"extensions"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Problem struct {
	Path        []interface{} `json:"path"`
	Explanation string        `json:"explanation"`
}

type Extensions struct {
	Value    interface{} `json:"value"`
	Problems []Problem   `json:"problems"`
}

type PullRequestNode struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type RepositoryNode struct {
	Name  string              `json:"name"`
	Owner RepositoryOwnerNode `json:"owner"`
}

type RepositoryOwnerNode struct {
	Login string `json:"login"`
}

func (g *GraphQLAPI) GetPullRequestByNodeID(ct context.Context, installation Installation, nodeID string) (PullRequestNode, *errs.Error) {
	queryOptions := gql.QueryOptions{
		Query: `
		query getPullRequest($nodeId:ID!) {
			node(id: $nodeId) {
				... on PullRequest {
					number
					repository {
						owner {
							login
						}
						name
					}
					url
				}
			}
		}`,
		Variables: struct {
			NodeID string
		}{
			NodeID: nodeID,
		},
	}

	var res gql.GraphQLResponse[Node[PullRequestNode], Error]
	err := g.query(ct, installation, queryOptions, &res)
	if err != nil {
		g.dataCollector.Logger.ErrorWithContext(ct, err)
		return PullRequestNode{}, err
	}

	if len(res.Errors) > 0 {
		var internalErr *errs.Error
		for _, resErr := range res.Errors {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: errors.New(resErr.Message),
			}
			g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		}

		return PullRequestNode{}, err
	}

	return res.Data.Node, nil
}

func (g *GraphQLAPI) query(
	ct context.Context,
	installation Installation,
	queryOptions gql.QueryOptions,
	gqlResponse interface{},
) *errs.Error {
	accessToken, err := installation.GetOrRefreshAccessToken(ct)
	if err != nil {
		g.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	headers := withCredential(make(map[string]string), accessToken)
	return g.graphQLClient.Query(ct, githubGraphQLAPIEndpoint, headers, queryOptions, gqlResponse)
}

func NewGraphQLAPI(dataCollector telemetry.DataCollector, graphQLClient gql.Client) *GraphQLAPI {
	return &GraphQLAPI{
		dataCollector: dataCollector,
		graphQLClient: graphQLClient,
	}
}
