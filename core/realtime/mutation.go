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
	Execute(ct context.Context, tx *transaction.Transaction) *errs.Error
	Undo() *errs.Error
	CleanUp(ct context.Context) *errs.Error
	GetClientNotifiers() []*ClientNotifier
	PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error
	ToMessage() MutationMessage
}
