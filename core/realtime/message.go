package realtime

type DeleteTaskAwaitForRelationPayload struct {
	AwaitingTaskID uint64
	AwaitForTaskID uint64
}

type DeleteSprintTaskRelationPayload struct {
	SprintID uint64
	TaskID   uint64
}

type MutationType string

const (
	CreateMutationType MutationType = "Create"
	UpdateMutationType MutationType = "Update"
	DeleteMutationType MutationType = "Delete"
)

type MessageType string

const (
	MutationMessageType MessageType = "Mutation"
	MetadataMessageType MessageType = "Metadata"
)

type MessageEvent struct {
	Type    MessageType
	Payload interface{}
}

type Mutation struct {
	ID             uint64
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
	TeamIDs        []uint64
}

type MutationMessage struct {
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
}

type MetadataMessage struct {
	ClientID uint64
}
