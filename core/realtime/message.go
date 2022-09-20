package realtime

type MessageType string

const (
	MutationMessageType MessageType = "Mutation"
	MetadataMessageType MessageType = "Metadata"
)

type Message struct {
	Type    MessageType
	Payload interface{}
}

type MutationMessage struct {
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
}

type MetadataMessage struct {
	ClientID uint64
}
