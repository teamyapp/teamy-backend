package realtime

type MutationType string

const (
	CreateMutationType MutationType = "Create"
	UpdateMutationType MutationType = "Update"
	DeleteMutationType MutationType = "Delete"
)

type DeleteTaskAwaitForRelationPayload struct {
	AwaitingTaskID uint64
	AwaitForTaskID uint64
}
