package entity

import (
	"github.com/teamyapp/cloud/app/lang"
)

type File struct {
	Metadata     FileMetadata
	ChunksBuffer chan lang.Result[[]byte]
}
