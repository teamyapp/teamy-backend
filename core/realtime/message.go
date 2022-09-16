package realtime

import "github.com/teamyapp/teamy-backend/core/entity"

type DeleteTaskAwaitForRelationPayload struct {
	AwaitingTaskID uint64
	AwaitForTaskID uint64
}

type DeleteSprintTaskRelationPayload struct {
	SprintID uint64
	TaskID   uint64
}

type Mutation struct {
	CollectionType entity.CollectionType
	MutationType   entity.MutationType
	Payload        interface{}
}

type Metadata struct {
	ClientID uint64
}
