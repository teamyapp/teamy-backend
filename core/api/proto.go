package api

import (
	"time"

	"github.com/teamyapp/cloud/libs/errs"
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

var protoPhaseStatuses = map[entity.PhaseStatus]message.PhaseStatus{
	entity.TodoPhaseStatus:       message.PhaseStatus_PHASE_TODO,
	entity.InProgressPhaseStatus: message.PhaseStatus_PHASE_IN_PROGRESS,
	entity.PausedPhaseStatus:     message.PhaseStatus_PHASE_PAUSED,
	entity.CompletedPhaseStatus:  message.PhaseStatus_PHASE_COMPLETED,
}

var protoStoryStatuses = map[entity.StoryStatus]message.StoryStatus{
	entity.TodoStoryStatus:       message.StoryStatus_STORY_TODO,
	entity.InProgressStoryStatus: message.StoryStatus_STORY_IN_PROGRESS,
	entity.PausedStoryStatus:     message.StoryStatus_STORY_PAUSED,
	entity.CompletedStoryStatus:  message.StoryStatus_STORY_COMPLETED,
}

var phaseStatuses = map[message.PhaseStatus]entity.PhaseStatus{
	message.PhaseStatus_PHASE_TODO:        entity.TodoPhaseStatus,
	message.PhaseStatus_PHASE_IN_PROGRESS: entity.InProgressPhaseStatus,
	message.PhaseStatus_PHASE_PAUSED:      entity.PausedPhaseStatus,
	message.PhaseStatus_PHASE_COMPLETED:   entity.CompletedPhaseStatus,
}

var storyStatuses = map[message.StoryStatus]entity.StoryStatus{
	message.StoryStatus_STORY_TODO:        entity.TodoStoryStatus,
	message.StoryStatus_STORY_IN_PROGRESS: entity.InProgressStoryStatus,
	message.StoryStatus_STORY_PAUSED:      entity.PausedStoryStatus,
	message.StoryStatus_STORY_COMPLETED:   entity.CompletedStoryStatus,
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

func fromProtoTime(ts *timestamppb.Timestamp) (time.Time, *errs.Error) {
	if ts == nil {
		return time.Time{}, errs.NewError(errs.InvalidArgument, "time is nil")
	}

	return ts.AsTime(), nil
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

func fromProtoPhaseStatusPtr(protoPhaseStatus *message.PhaseStatus) *entity.PhaseStatus {
	if protoPhaseStatus == nil {
		return nil
	}

	phaseStatus := phaseStatuses[*protoPhaseStatus]
	return &phaseStatus
}

func fromProtoStoryStatusPtr(protoStoryStatus *message.StoryStatus) *entity.StoryStatus {
	if protoStoryStatus == nil {
		return nil
	}

	storyStatus := storyStatuses[*protoStoryStatus]
	return &storyStatus
}
