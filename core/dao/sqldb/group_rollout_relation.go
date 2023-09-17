package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupRolloutRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.GroupRolloutRelation = (*GroupRolloutRelation)(nil)

func (g *GroupRolloutRelation) FindGroupRolloutRelationsByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error) {
	groupRolloutRelations := []entity.GroupRolloutRelation{}
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			group_id,
			rollout_id,
			order_index
			FROM group_rollout_relations
			WHERE group_id = $1
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

func (g *GroupRolloutRelation) CreateGroupRolloutRelation(ct context.Context, tx *transaction.Transaction, groupRolloutRelation entity.GroupRolloutRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`INSERT INTO group_rollout_relations (
			group_id,
			rollout_id,
			order_index
		) VALUES (
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

func NewGroupRolloutRelation(
	transactionFactory transaction.Factory,
) *GroupRolloutRelation {
	return &GroupRolloutRelation{
		transactionFactory: transactionFactory,
	}
}
