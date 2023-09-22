package api

import (
	"time"

	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var protoTaskStatuses = map[entity.TaskStatus]proto.TaskStatus{
	entity.TaskStatusTodo:       proto.TaskStatus_Todo,
	entity.TaskStatusInProgress: proto.TaskStatus_InProgress,
	entity.TaskStatusPaused:     proto.TaskStatus_Paused,
	entity.TaskStatusAwaiting:   proto.TaskStatus_Awaiting,
	entity.TaskStatusBlocked:    proto.TaskStatus_Blocked,
	entity.TaskStatusDelivered:  proto.TaskStatus_Delivered,
}

var protoPriorities = map[entity.Priority]proto.Priority{
	entity.UrgentPriority: proto.Priority_Urgent,
	entity.HighPriority:   proto.Priority_High,
	entity.MediumPriority: proto.Priority_Medium,
	entity.LowPriority:    proto.Priority_Low,
}

var priorities = map[proto.Priority]entity.Priority{
	proto.Priority_Urgent: entity.UrgentPriority,
	proto.Priority_High:   entity.HighPriority,
	proto.Priority_Medium: entity.MediumPriority,
	proto.Priority_Low:    entity.LowPriority,
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

func fromProtoDurationPtr(protoDuration *durationpb.Duration) *time.Duration {
	if protoDuration == nil {
		return nil
	}

	dur := protoDuration.AsDuration()
	return &dur
}

func toProtoDurationPtr(duration *time.Duration) *durationpb.Duration {
	if duration == nil {
		return nil
	}

	return durationpb.New(*duration)
}

func toProtoPriorityPtr(priority *entity.Priority) *proto.Priority {
	if priority == nil {
		return nil
	}

	protoPriority := protoPriorities[*priority]
	return &protoPriority
}

func fromProtoPriorityPtr(protoPriority *proto.Priority) *entity.Priority {
	if protoPriority == nil {
		return nil
	}

	priority := priorities[*protoPriority]
	return &priority
}
