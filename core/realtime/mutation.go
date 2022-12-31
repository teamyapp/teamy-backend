package realtime

import (
	"context"
)

type MutationType string

const (
	CreateMutationType MutationType = "Create"
	UpdateMutationType MutationType = "Update"
	DeleteMutationType MutationType = "Delete"
)

type Mutation interface {
	CleanUp(ct context.Context) error
	GetID() uint64
	Execute(ct context.Context) error
	Undo() error
	GetClientNotifiers(ct context.Context) ([]*ClientNotifier, error)
	ToMessage() MutationMessage
}
