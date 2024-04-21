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

const groupDaoName = "Group"

type Group struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.Group = (*Group)(nil)

func (g *Group) FindGroupByIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (entity.Group, *errs.Error) {
	g.metrics.ReportDaoOperation(groupDaoName, "FindGroupByIDWithTx")
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
			locked,
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
		&group.Locked,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err != nil {
		return entity.Group{}, errs.NewError(errs.Unknown, err.Error())
	}

	return group, nil
}

func (g *Group) FindGroupByID(ct context.Context, groupID uint64) (entity.Group, *errs.Error) {
	g.metrics.ReportDaoOperation(groupDaoName, "FindGroupByID")
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
	g.metrics.ReportDaoOperation(groupDaoName, "FindGroupsByIDsWithTx")
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
			locked,
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
			&group.Locked,
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
	g.metrics.ReportDaoOperation(groupDaoName, "FindGroupsByIDs")
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
	g.metrics.ReportDaoOperation(groupDaoName, "CreateGroup")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO "group" (
			id,
			type,
			member_type,
			max_rollout_index,
			name,
			locked,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
		`,
		group.ID,
		group.Type,
		group.MemberType,
		group.MaxRolloutIndex,
		group.Name,
		group.Locked,
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *Group) UpdateGroup(ct context.Context, tx *transaction.Transaction, Group entity.Group) *errs.Error {
	g.metrics.ReportDaoOperation(groupDaoName, "UpdateGroup")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		UPDATE "group"
		SET
		    type = $1,
		    member_type = $2,
		    max_rollout_index = $3,
		    name = $4,
			locked = $5,
		    created_at = $6,
		    updated_at = $7
		WHERE id = $8;
		`,
		Group.Type,
		Group.MemberType,
		Group.MaxRolloutIndex,
		Group.Name,
		Group.Locked,
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
	g.metrics.ReportDaoOperation(groupDaoName, "DeleteGroup")
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

func (g *Group) FilterGroupIDsByMemberTypeWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	groupIDs []uint64,
	memberType entity.GroupMemberType,
) ([]uint64, *errs.Error) {
	g.metrics.ReportDaoOperation(groupDaoName, "FilterGroupIDsByMemberTypeWithTx")
	if len(groupIDs) == 0 {
		return []uint64{}, nil
	}

	idsString := toIDsString(groupIDs)
	query := fmt.Sprintf(
		`
		SELECT
			id
		FROM "group"
		WHERE id IN (%s) AND member_type = $1;
		`,
		idsString,
	)

	rows, err := tx.SQLTx().QueryContext(ct, query, memberType)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	result := make([]uint64, 0)
	for rows.Next() {
		var id uint64
		err := rows.Scan(&id)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		result = append(result, id)
	}

	return result, nil
}

func NewGroup(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *Group {
	return &Group{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
