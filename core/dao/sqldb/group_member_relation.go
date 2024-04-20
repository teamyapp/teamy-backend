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

const groupMemberRelationDaoName = "GroupMemberRelation"

type GroupMemberRelation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.GroupMemberRelation = (*GroupMemberRelation)(nil)

func (g *GroupMemberRelation) FindMemberIDsByGroupIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) ([]uint64, *errs.Error) {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "FindMemberIDsByGroupIDWithTx")
	var memberIDs []uint64
	row, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT member_id
		FROM group_member_relation
		WHERE group_id = $1`,
		groupID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer row.Close()

	for row.Next() {
		var memberID uint64
		err := row.Scan(&memberID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		memberIDs = append(memberIDs, memberID)
	}

	return memberIDs, nil
}

func (g *GroupMemberRelation) FindMemberIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "FindMemberIDsByGroupID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := g.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return g.FindMemberIDsByGroupIDWithTx(ct, tx, groupID)
}

func (g *GroupMemberRelation) FilterGroupIDsByMemberID(ct context.Context, groupIDs []uint64, memberID uint64) ([]uint64, *errs.Error) {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "FilterGroupIDsByMemberID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := g.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return g.FilterGroupIDsByMemberIDWithTx(ct, tx, groupIDs, memberID)
}

func (g *GroupMemberRelation) FilterGroupIDsByMemberIDWithTx(ct context.Context, tx *transaction.Transaction, groupIDs []uint64, memberID uint64) ([]uint64, *errs.Error) {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "FilterGroupIDsByMemberIDWithTx")
	groupIDsString := toIDsString(groupIDs)
	query := fmt.Sprintf(`
		SELECT group_id
		FROM group_member_relation
		WHERE group_id IN (%s) AND member_id = $1`,
		groupIDsString,
	)
	row, err := tx.SQLTx().QueryContext(
		ct,
		query,
		memberID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer row.Close()

	groupIDs = []uint64{}
	for row.Next() {
		var groupID uint64
		err := row.Scan(&groupID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs, nil
}

func (g *GroupMemberRelation) CreateGroupMemberRelation(
	ct context.Context,
	tx *transaction.Transaction,
	groupMemberRelation entity.GroupMemberRelation,
) *errs.Error {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "CreateGroupMemberRelation")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO group_member_relation (
			member_id,
			group_id
		)
		VALUES ($1, $2);`,
		groupMemberRelation.MemberID,
		groupMemberRelation.GroupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *GroupMemberRelation) DeleteGroupMemberRelation(ct context.Context, tx *transaction.Transaction, memberID uint64, groupID uint64) *errs.Error {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "DeleteGroupMemberRelation")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		DELETE FROM group_member_relation
		WHERE member_id = $1 AND group_id = $2;`,
		memberID,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g *GroupMemberRelation) DeleteGroupMemberRelationsByGroupID(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error {
	g.metrics.ReportDaoOperation(groupMemberRelationDaoName, "DeleteGroupMemberRelationsByGroupID")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		DELETE FROM group_member_relation
		WHERE group_id = $1;`,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewGroupMemberRelation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *GroupMemberRelation {
	return &GroupMemberRelation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
