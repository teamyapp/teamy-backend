package daotestv2

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.User = (*User)(nil)

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := u.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.User{}, err
	}

	defer tx.Rollback()
	return u.FindUserByIDWithTx(ct, tx, userID)
}

func (u User) FindUserByIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64) (entity.User, *errs.Error) {
	var user entity.User
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := u.db.GetTable(UserTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUser := rawRow.(entity.User)
				if currUser.ID == userID {
					user = currUser
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: userID=%v", userID),
			}
		},
	})
	return user, err
}

func (u User) FindUsersByIDsWithTx(ct context.Context, tx *transaction.Transaction, userIDs []uint64) ([]entity.User, *errs.Error) {
	var users []entity.User
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := u.db.GetTable(UserTableName)
			if err != nil {
				return err
			}

			userMap := make(map[uint64]int)
			for _, userID := range userIDs {
				userMap[userID]++
			}

			for _, rawRow := range table.Rows {
				currUser := rawRow.(entity.User)
				if _, ok := userMap[currUser.ID]; ok {
					users = append(users, currUser)
				}
			}

			return nil
		},
	})
	return users, err
}

func (u User) CreateUser(ct context.Context, tx *transaction.Transaction, user entity.User) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := u.db.GetTable(UserTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currUser := row.(entity.User)
				if currUser.ID == user.ID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: userID=%v", user.ID))
				}
			}

			table.Rows = append(table.Rows, user)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := u.db.GetTable(UserTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currUser := row.(entity.User)
				if currUser.ID == user.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (u User) UpdateUser(ct context.Context, tx *transaction.Transaction, user entity.User) *errs.Error {
	oldUser, internalErr := u.FindUserByIDWithTx(ct, tx, user.ID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := u.db.GetTable(UserTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currUser := row.(entity.User)
				if currUser.ID == user.ID {
					table.Rows[i] = user
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: userID=%v", user.ID))
		},
		Undo: func() *errs.Error {
			table, err := u.db.GetTable(UserTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currUser := row.(entity.User)
				if currUser.ID == user.ID {
					table.Rows[index] = oldUser
				}
			}

			return nil
		},
	})
}

func NewUser(db *dbtest.InMemoryDB) User {
	return User{db: db}
}
