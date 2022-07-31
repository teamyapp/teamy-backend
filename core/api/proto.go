package api

import (
	"time"

	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var protoTaskStatuses = map[entity.TaskStatus]proto.TaskStatus{
	entity.TaskStatusUpcoming:   proto.TaskStatus_Upcoming,
	entity.TaskStatusInProgress: proto.TaskStatus_InProgress,
	entity.TaskStatusPaused:     proto.TaskStatus_Paused,
	entity.TaskStatusAwaiting:   proto.TaskStatus_Awaiting,
	entity.TaskStatusBlocked:    proto.TaskStatus_Blocked,
	entity.TaskStatusDelivered:  proto.TaskStatus_Delivered,
}

func fromProtoTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	tm := ts.AsTime()
	return &tm
}

func toProtoTimePtr(tm *time.Time) *timestamppb.Timestamp {
	if tm == nil {
		return nil
	}

	return timestamppb.New(*tm)
}

func fromProtoInt32Ptr(num *int32) *int {
	if num == nil {
		return nil
	}

	newNum := int(*num)
	return &newNum
}

func toProtoInt32Ptr(num *int) *int32 {
	if num == nil {
		return nil
	}

	newNum := int32(*num)
	return &newNum
}
