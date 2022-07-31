package api

import (
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var toProtoUploadSessionStatus = map[entity.UploadSessionStatus]proto.UploadSessionStatus{
	entity.CreatedUploadSessionStatus:         proto.UploadSessionStatus_CREATED,
	entity.InitializedUploadSessionStatus:     proto.UploadSessionStatus_INITIALIZED,
	entity.UploadingChunksUploadSessionStatus: proto.UploadSessionStatus_UPLOADING_CHUNKS,
	entity.CompletedUploadSessionStatus:       proto.UploadSessionStatus_COMPLETED,
}

func toProtoTimePtr(tm *time.Time) *timestamppb.Timestamp {
	if tm == nil {
		return nil
	}

	return timestamppb.New(*tm)
}
