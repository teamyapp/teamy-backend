package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type FileMetadata struct {
	deps         *Dependencies
	fileMetadata entity.FileMetadata
}

func (f FileMetadata) ID(ct context.Context) graphql.ID {
	return toGraphQLID(f.fileMetadata.ID)
}

func (f FileMetadata) Name(ct context.Context) string {
	return f.fileMetadata.Name
}

func (f FileMetadata) OwningTeam(ct context.Context) (Team, error) {
	panic("implement me")
}

func (f FileMetadata) CreatedAt(ct context.Context) graphql.Time {
	return graphql.Time{Time: f.fileMetadata.CreatedAt}
}

func (f FileMetadata) LastModifiedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(f.fileMetadata.LastModifiedAt)
}

func newFileMetadata(deps *Dependencies, fileMetadata entity.FileMetadata) FileMetadata {
	return FileMetadata{
		deps:         deps,
		fileMetadata: fileMetadata,
	}
}
