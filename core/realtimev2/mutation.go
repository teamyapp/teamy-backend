package realtimev2

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
	Execute(ct context.Context, sqlTx *sql.Tx, rtTx *Transaction) *errs.Error
	CleanUp(ct context.Context) *errs.Error
	GetClientNotifiers(ct context.Context) ([]*ClientNotifier, *errs.Error)
	ToMessage() MutationMessage
}
