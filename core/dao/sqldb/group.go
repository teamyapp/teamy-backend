package sqldb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group struct {
	transactionFactory transaction.Factory
}

var _ dao.Group = (*Group)(nil)

func (*Group) FindGroupByIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (entity.Group, *errs.Error) {
	group := entity.Group{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			id,
			type,
			member_type,
			max_rollout_index,
			name,
			created_at,
			updated_at
		FROM "group"
		WHERE id = $1`,
		groupID,
	).Scan(
		&group.ID,
		&group.Type,
		&group.MemberType,
		&group.MaxRolloutIndex,
		&group.Name,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err != nil {
		return entity.Group{}, errs.NewError(errs.Unknown, err.Error())
	}

	return group, nil
}

func (g *Group) FindGroupByID(ct context.Context, groupID uint64) (entity.Group, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := g.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Group{}, err
	}

	defer tx.Rollback()
	return g.FindGroupByIDWithTx(ct, tx, groupID)
}

func (g *Group) FindGroupsByIDsWithTx(ct context.Context, tx *transaction.Transaction, groupIDs []uint64) ([]entity.Group, *errs.Error) {
	if len(groupIDs) == 0 {
		return []entity.Group{}, nil
	}

	groups := make([]entity.Group, 0)
	idsString := toIDsString(groupIDs)
	query := fmt.Sprintf(
		`
		SELECT
			id,
			type,
			member_type,
			max_rollout_index,
			name,
			created_at,
			updated_at
		FROM "group"
		WHERE id IN (%s);
		`,
		idsString,
	)

	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		group := entity.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Type,
			&group.MemberType,
			&group.MaxRolloutIndex,
			&group.Name,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (g *Group) FindGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.Group, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := g.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return g.FindGroupsByIDsWithTx(ct, tx, groupIDs)
}

func (g *Group) CreateGroup(ct context.Context, tx *transaction.Transaction, group entity.Group) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO "group" (
			id,
			type,
			member_type,
			max_rollout_index,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
		`,
		group.ID,
		group.Type,
		group.MemberType,
		group.MaxRolloutIndex,
		group.Name,
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*Group) UpdateGroup(ct context.Context, tx *transaction.Transaction, Group entity.Group) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		UPDATE "group"
		SET
		    type = $1,
			member_type = $2,
			max_rollout_index = $3,
		    name = $4,
			created_at = $5,
		    updated_at = $6
		WHERE id = $7;
		`,
		Group.Type,
		Group.MemberType,
		Group.MaxRolloutIndex,
		Group.Name,
		Group.CreatedAt,
		Group.UpdatedAt,
		Group.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *Group) DeleteGroup(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM "group"
		WHERE id = $1;
		`,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewGroup(transactionFactory transaction.Factory) *Group {
	return &Group{
		transactionFactory: transactionFactory,
	}
}
