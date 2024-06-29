package api

import (
	"time"

	"github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var attachmentListOwnerTypes = map[entity.AttachmentListOwnerType]message.AttachmentListOwnerType{
	entity.AttachmentListOwnerTypeTask: message.AttachmentListOwnerType_TASK,
}

var protoAttachmentListOwnerTypes = map[message.AttachmentListOwnerType]entity.AttachmentListOwnerType{
	message.AttachmentListOwnerType_TASK: entity.AttachmentListOwnerTypeTask,
}

var attachmentTypes = map[entity.AttachmentType]message.AttachmentType{
	entity.AttachmentTypeImage: message.AttachmentType_IMAGE,
}

var protoTaskStatuses = map[entity.TaskStatus]message.TaskStatus{
	entity.TaskStatusTodo:       message.TaskStatus_TODO,
	entity.TaskStatusInProgress: message.TaskStatus_IN_PROGRESS,
	entity.TaskStatusPaused:     message.TaskStatus_PAUSED,
	entity.TaskStatusAwaiting:   message.TaskStatus_AWAITING,
	entity.TaskStatusBlocked:    message.TaskStatus_BLOCKED,
	entity.TaskStatusDelivered:  message.TaskStatus_DELIVERED,
}

var protoPriorities = map[entity.Priority]message.Priority{
	entity.UrgentPriority: message.Priority_URGENT,
	entity.HighPriority:   message.Priority_HIGH,
	entity.MediumPriority: message.Priority_MEDIUM,
	entity.LowPriority:    message.Priority_LOW,
}

var priorities = map[message.Priority]entity.Priority{
	message.Priority_URGENT: entity.UrgentPriority,
	message.Priority_HIGH:   entity.HighPriority,
	message.Priority_MEDIUM: entity.MediumPriority,
	message.Priority_LOW:    entity.LowPriority,
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

func toProtoPriorityPtr(priority *entity.Priority) *message.Priority {
	if priority == nil {
		return nil
	}

	protoPriority := protoPriorities[*priority]
	return &protoPriority
}

func fromProtoPriorityPtr(protoPriority *message.Priority) *entity.Priority {
	if protoPriority == nil {
		return nil
	}

	priority := priorities[*protoPriority]
	return &priority
}
