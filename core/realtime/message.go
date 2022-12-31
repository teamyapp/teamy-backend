package realtime

type MessageType string

const (
	MetadataMessageType    MessageType = "Metadata"
	MutationMessageType    MessageType = "Mutation"
	TransactionMessageType MessageType = "Transaction"
)

type Message struct {
	Type    MessageType
	Payload interface{}
}

type MetadataMessage struct {
	ClientID uint64
}

type MutationMessage struct {
	ID             uint64
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
}

type TransactionMessage struct {
	ID        uint64
	Mutations []MutationMessage
}
