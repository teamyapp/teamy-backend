package realtime

type MutationType string

const (
	CreateMutationType MutationType = "Create"
	UpdateMutationType MutationType = "Update"
	DeleteMutationType MutationType = "Delete"
)

type Mutation struct {
	ID             uint64
	CollectionType CollectionType
	MutationType   MutationType
	Payload        interface{}
	TeamIDs        []uint64
}
