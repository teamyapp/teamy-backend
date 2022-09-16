package entity

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
	ID      uint64
	Payload interface{}
}

type MutationPayload struct {
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
	TeamIDs        []uint64
}

type MetadataPayload struct {
	ClientID uint64
	UserID   uint64
}
