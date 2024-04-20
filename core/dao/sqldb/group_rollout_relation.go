package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const groupRolloutRelationDaoName = "GroupRolloutRelation"

type GroupRolloutRelation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.GroupRolloutRelation = (*GroupRolloutRelation)(nil)

func (g *GroupRolloutRelation) FindGroupRolloutRelationsByGroupIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) ([]entity.GroupRolloutRelation, *errs.Error) {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "FindGroupRolloutRelationsByGroupIDWithTx")
	var groupRolloutRelations []entity.GroupRolloutRelation
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			group_id,
			rollout_id,
			order_index
		FROM group_rollout_relation
		WHERE group_id = $1;
		`,
		groupID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		groupRolloutRelation := entity.GroupRolloutRelation{}
		err := rows.Scan(
			&groupRolloutRelation.GroupID,
			&groupRolloutRelation.RolloutID,
			&groupRolloutRelation.OrderIndex,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		groupRolloutRelations = append(groupRolloutRelations, groupRolloutRelation)
	}

	return groupRolloutRelations, nil
}

func (g *GroupRolloutRelation) FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error) {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "FindGroupRolloutRelationsByGroupID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := g.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return g.FindGroupRolloutRelationsByGroupIDWithTx(ct, tx, groupID)
}

func (g *GroupRolloutRelation) FindGroupRolloutByGroupIDAndRolloutIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	groupID,
	rolloutID uint64,
) (*entity.GroupRolloutRelation, *errs.Error) {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "FindGroupRolloutByGroupIDAndRolloutIDWithTx")
	groupRolloutRelation := entity.GroupRolloutRelation{}
	row := tx.SQLTx().QueryRowContext(ct,
		`
		SELECT
			group_id,
			rollout_id,
			order_index
		FROM group_rollout_relation
		WHERE group_id = $1
		AND rollout_id = $2;
		`,
		groupID,
		rolloutID,
	)

	err := row.Scan(
		&groupRolloutRelation.GroupID,
		&groupRolloutRelation.RolloutID,
		&groupRolloutRelation.OrderIndex,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return &groupRolloutRelation, nil
}

func (g *GroupRolloutRelation) FindGroupRolloutRelationsByRolloutIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutID uint64,
) ([]entity.GroupRolloutRelation, *errs.Error) {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "FindGroupRolloutRelationsByRolloutIDWithTx")
	var groupRolloutRelations []entity.GroupRolloutRelation
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			group_id,
			rollout_id,
			order_index
		FROM group_rollout_relation
		WHERE rollout_id = $1;
		`,
		rolloutID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		groupRolloutRelation := entity.GroupRolloutRelation{}
		err := rows.Scan(
			&groupRolloutRelation.GroupID,
			&groupRolloutRelation.RolloutID,
			&groupRolloutRelation.OrderIndex,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		groupRolloutRelations = append(groupRolloutRelations, groupRolloutRelation)
	}

	return groupRolloutRelations, nil
}

func (g *GroupRolloutRelation) FindGroupRolloutRelationsByRolloutID(ct context.Context, rolloutID uint64) ([]entity.GroupRolloutRelation, *errs.Error) {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "FindGroupRolloutRelationsByRolloutID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := g.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return g.FindGroupRolloutRelationsByRolloutIDWithTx(ct, tx, rolloutID)
}

func (g *GroupRolloutRelation) CreateGroupRolloutRelation(ct context.Context, tx *transaction.Transaction, groupRolloutRelation entity.GroupRolloutRelation) *errs.Error {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "CreateGroupRolloutRelation")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO group_rollout_relation (
			group_id,
			rollout_id,
			order_index
		)
		VALUES (
			$1,
			$2,
			$3
		)`,
		groupRolloutRelation.GroupID,
		groupRolloutRelation.RolloutID,
		groupRolloutRelation.OrderIndex,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *GroupRolloutRelation) DeleteGroupRolloutRelationsByGroupID(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "DeleteGroupRolloutRelationsByGroupID")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM group_rollout_relation
		WHERE group_id = $1;`,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *GroupRolloutRelation) DeleteGroupRolloutRelationsByGroupIDAndRolloutID(ct context.Context, tx *transaction.Transaction, groupID, rolloutID uint64) *errs.Error {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "DeleteGroupRolloutRelationsByGroupIDAndRolloutID")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM group_rollout_relation
		WHERE group_id = $1
		AND rollout_id = $2;`,
		groupID,
		rolloutID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *GroupRolloutRelation) DeleteGroupRolloutRelationsByRolloutID(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutID uint64,
) *errs.Error {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "DeleteGroupRolloutRelationsByRolloutID")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM group_rollout_relation
		WHERE rollout_id = $1;`,
		rolloutID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *GroupRolloutRelation) UpdateFromOrderIndexByGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	step int,
	orderIndex int,
	groupID uint64,
) *errs.Error {
	g.metrics.ReportDaoOperation(groupRolloutRelationDaoName, "UpdateFromOrderIndexByGroupID")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		UPDATE group_rollout_relation
		SET order_index = order_index + $1
		WHERE group_id = $2 AND order_index >= $3;`,
		step,
		groupID,
		orderIndex,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewGroupRolloutRelation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *GroupRolloutRelation {
	return &GroupRolloutRelation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
