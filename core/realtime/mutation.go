package realtime

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
)

type MutationType string

const (
	CreateMutationType MutationType = "Create"
	UpdateMutationType MutationType = "Update"
	DeleteMutationType MutationType = "Delete"
)

type Mutation interface {
	GetID() uint64
	// Deprecated: The old method should not be used anymore. Use ExecuteV2 method instead
	Execute(ct context.Context) *errs.Error
	ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error
	Undo() *errs.Error
	CleanUp(ct context.Context) *errs.Error
	// Deprecated: The old method should not be used anymore. Use another PrepareClientNotifiers method instead
	GetClientNotifiers(ct context.Context) ([]*ClientNotifier, *errs.Error)
	PrepareClientNotifiers(ct context.Context, tx *sql.Tx) ([]*ClientNotifier, *errs.Error)
	ToMessage() MutationMessage
}
