package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const userDaoName = "User"

type User struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.User = (*User)(nil)

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	u.metrics.ReportDaoOperation(userDaoName, "FindUserByID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := u.transactionFactory.BeginTx(ct, &opt)
	defer tx.Rollback()
	if err != nil {
		return entity.User{}, err
	}

	return u.FindUserByIDWithTx(ct, tx, userID)
}

func (u User) FindUserByIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64) (entity.User, *errs.Error) {
	u.metrics.ReportDaoOperation(userDaoName, "FindUserByIDWithTx")
	statement := `
	SELECT
		id,
		first_name,
		last_name,
		profile_url,
		created_at,
		updated_at
	FROM "user"
	WHERE id = $1;
`
	user := entity.User{}
	err := tx.SQLTx().QueryRow(statement, userID).
		Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.ProfileURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, errs.NewError(errs.NotFound, fmt.Sprintf("user not found: userID=%v", userID))
	}

	if err != nil {
		return entity.User{}, errs.NewError(errs.Unknown, err.Error())
	}

	return user, nil
}

func (u User) FindUsersByIDsWithTx(ct context.Context, tx *transaction.Transaction, userIDs []uint64) ([]entity.User, *errs.Error) {
	u.metrics.ReportDaoOperation(userDaoName, "FindUsersByIDsWithTx")
	if len(userIDs) == 0 {
		return nil, nil
	}

	idsString := toIDsString(userIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		first_name,
		last_name,
		profile_url,
		created_at,
		updated_at
	FROM "user"
	WHERE id IN (%v)`, idsString)
	rows, err := tx.SQLTx().Query(query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var internalErr *errs.Error
	var users []entity.User
	for rows.Next() {
		var user entity.User
		err = rows.
			Scan(
				&user.ID,
				&user.FirstName,
				&user.LastName,
				&user.ProfileURL,
				&user.CreatedAt,
				&user.UpdatedAt,
			)
		if err != nil {
			if internalErr == nil {
				internalErr = errs.NewError(errs.Unknown, err.Error())
			}

			continue
		}

		users = append(users, user)
	}

	return users, internalErr
}

func (u User) CreateUser(ct context.Context, tx *transaction.Transaction, user entity.User) *errs.Error {
	u.metrics.ReportDaoOperation(userDaoName, "CreateUser")
	if tx.SQLTx() == nil {
		panic("It's nil")
	}

	_, err := tx.SQLTx().Exec(`
		INSERT INTO "user"
		(
			id,
			first_name,
			last_name,
			profile_url,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5);`,
		user.ID,
		user.FirstName,
		user.LastName,
		user.ProfileURL,
		user.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (u User) UpdateUser(ct context.Context, tx *transaction.Transaction, user entity.User) *errs.Error {
	u.metrics.ReportDaoOperation(userDaoName, "UpdateUser")
	_, err := tx.SQLTx().Exec(`
		UPDATE "user"
		SET
			first_name = $1,
			last_name = $2,
			profile_url = $3,
			updated_at = $4
		WHERE id = $5;`,
		user.FirstName,
		user.LastName,
		user.ProfileURL,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewUser(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) User {
	return User{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
