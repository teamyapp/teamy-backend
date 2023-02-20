package client

import (
	"context"
	"fmt"

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

type Error struct {
	// TODO(szheng2207): explore Github GraphQL API error structure
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
		// TODO(szheng2207): handle multiple errors from Github
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: fmt.Errorf("%+v", res.Errors[0]),
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
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
