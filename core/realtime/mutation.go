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
	GetID() uint64
	Execute(ct context.Context) error
	Undo() error
	CleanUp(ct context.Context) error
	GetClientNotifiers(ct context.Context) ([]*ClientNotifier, error)
	ToMessage() MutationMessage
}
