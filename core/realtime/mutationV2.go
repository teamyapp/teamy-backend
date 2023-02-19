package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type MutationTypeV2 string

const (
	CreateMutationTypeV2 MutationType = "Create"
	UpdateMutationTypeV2 MutationType = "Update"
	DeleteMutationTypeV2 MutationType = "Delete"
)

type MutationV2 interface {
	GetID() uint64
	Execute(ct context.Context) *errs.Error
	CleanUp(ct context.Context) *errs.Error
	GetClientNotifiers(ct context.Context) ([]*ClientNotifier, *errs.Error)
	ToMessage() MutationMessage
}
