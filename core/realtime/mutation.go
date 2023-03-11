package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
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
	ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error
	Undo() *errs.Error
	CleanUp(ct context.Context) *errs.Error
	// Deprecated: use PrepareClientNotifiers and GetClientNotifiersV2 method instead
	GetClientNotifiers(ct context.Context) ([]*ClientNotifier, *errs.Error)
	GetClientNotifiersV2() []*ClientNotifier
	PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error
	ToMessage() MutationMessage
}
