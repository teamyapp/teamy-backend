package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User interface {
	FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error)
	FindUserByIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64) (entity.User, *errs.Error)
	FindUsersByIDsWithTx(ct context.Context, tx *transaction.Transaction, userIDs []uint64) ([]entity.User, *errs.Error)
	CreateUser(ct context.Context, tx *transaction.Transaction, user entity.User) *errs.Error
	UpdateUser(ct context.Context, tx *transaction.Transaction, user entity.User) *errs.Error
}
