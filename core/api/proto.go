package api

import (
	"time"

	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var protoTaskStatuses = map[entity.TaskStatus]pbteamy.TaskStatus{
	entity.TaskStatusTodo:       pbteamy.TaskStatus_Todo,
	entity.TaskStatusInProgress: pbteamy.TaskStatus_InProgress,
	entity.TaskStatusPaused:     pbteamy.TaskStatus_Paused,
	entity.TaskStatusAwaiting:   pbteamy.TaskStatus_Awaiting,
	entity.TaskStatusBlocked:    pbteamy.TaskStatus_Blocked,
	entity.TaskStatusDelivered:  pbteamy.TaskStatus_Delivered,
}

var protoPriorities = map[entity.Priority]pbteamy.Priority{
	entity.UrgentPriority: pbteamy.Priority_Urgent,
	entity.HighPriority:   pbteamy.Priority_High,
	entity.MediumPriority: pbteamy.Priority_Medium,
	entity.LowPriority:    pbteamy.Priority_Low,
}

var priorities = map[pbteamy.Priority]entity.Priority{
	pbteamy.Priority_Urgent: entity.UrgentPriority,
	pbteamy.Priority_High:   entity.HighPriority,
	pbteamy.Priority_Medium: entity.MediumPriority,
	pbteamy.Priority_Low:    entity.LowPriority,
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

func toProtoPriorityPtr(priority *entity.Priority) *pbteamy.Priority {
	if priority == nil {
		return nil
	}

	protoPriority := protoPriorities[*priority]
	return &protoPriority
}

func fromProtoPriorityPtr(protoPriority *pbteamy.Priority) *entity.Priority {
	if protoPriority == nil {
		return nil
	}

	priority := priorities[*protoPriority]
	return &priority
}
