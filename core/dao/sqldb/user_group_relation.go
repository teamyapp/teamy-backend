package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserGroupRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.UserGroupRelation = (*UserGroupRelation)(nil)

func (*UserGroupRelation) FindUserIDsByGroupIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) ([]uint64, *errs.Error) {
	userIDs := make([]uint64, 0)
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`SELECT
			user_id
			FROM user_group_relation
			WHERE group_id = $1`,
		groupID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var userID uint64
		err := rows.Scan(&userID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (u *UserGroupRelation) FindUserIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := u.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return u.FindUserIDsByGroupIDWithTx(ct, tx, groupID)
}
func (*UserGroupRelation) CreateUserGroupRelation(
	ct context.Context,
	tx *transaction.Transaction,
	userGroupRelation entity.UserGroupRelation,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`INSERT INTO user_group_relation (
			user_id,
			group_id
		) VALUES (
			$1,
			$2
		)`,
		userGroupRelation.UserID,
		userGroupRelation.GroupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*UserGroupRelation) DeleteUserGroupRelation(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
	userID uint64,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`DELETE FROM user_group_relation
			WHERE group_id = $1 AND user_id = $2`,
		groupID,
		userID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewUserGroupRelation(
	transactionFactory transaction.Factory,
) *UserGroupRelation {
	return &UserGroupRelation{
		transactionFactory: transactionFactory,
	}
}
