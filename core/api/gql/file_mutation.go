package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

func (m Mutation) CreateFileUploadSession(ct context.Context, args struct {
	TeamID graphql.ID
}) (graphql.ID, error) {
	// TODO: implement me
	panic("implement me")
}

func (m Mutation) FinishFileUploadSession(ct context.Context, args struct {
	TeamID graphql.ID
}) (FileMetadata, error) {
	// TODO: implement me
	panic("implement me")
}

func (m Mutation) UpdateFileMetadata(ct context.Context, args struct {
	FileID graphql.ID
	Input  struct {
		Name string
	}
}) (FileMetadata, error) {
	// TODO: implement me
	panic("implement me")
}

func (m Mutation) DeleteFile(ct context.Context, args struct {
	FileID graphql.ID
}) (FileMetadata, error) {
	// TODO: implement me
	panic("implement me")
}
