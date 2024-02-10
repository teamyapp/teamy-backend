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

type FilterGroup struct {
	transactionFactory transaction.Factory
}

var _ dao.FilterGroup = (*FilterGroup)(nil)

func (f *FilterGroup) FindFilterGroupByIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (entity.FilterGroup, *errs.Error) {
	filterGroup := entity.FilterGroup{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
		SELECT
			group_id,
			filter
		FROM filter_group
		WHERE group_id = $1;`,
		groupID,
	).Scan(
		&filterGroup.Group.ID,
		&filterGroup.Filter,
	)

	if err != nil {
		return entity.FilterGroup{}, errs.NewError(errs.Unknown, err.Error())
	}

	return filterGroup, nil
}

func (f *FilterGroup) FindFilterGroupByID(ct context.Context, groupID uint64) (entity.FilterGroup, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := f.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	defer tx.Rollback()
	return f.FindFilterGroupByIDWithTx(ct, tx, groupID)
}

func (f *FilterGroup) FindFilterGroupsByIDsWithTx(ct context.Context, tx *transaction.Transaction, groupIDs []uint64) ([]entity.FilterGroup, *errs.Error) {
	if len(groupIDs) == 0 {
		return []entity.FilterGroup{}, nil
	}

	idsString := toIDsString(groupIDs)
	filterGroups := []entity.FilterGroup{}
	query := fmt.Sprintf(
		`
		SELECT
			group_id,
			filter,
			count
		FROM filter_group
		WHERE group_id IN (%s);`,
		idsString,
	)

	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		filterGroup := entity.FilterGroup{}
		err := rows.Scan(
			&filterGroup.Group.ID,
			&filterGroup.Filter,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		filterGroups = append(filterGroups, filterGroup)
	}

	return filterGroups, nil
}

func (f *FilterGroup) FindFilterGroupsByIDs(ct context.Context, groupID []uint64) ([]entity.FilterGroup, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := f.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return f.FindFilterGroupsByIDsWithTx(ct, tx, groupID)
}

func (f *FilterGroup) CreateFilterGroup(ct context.Context, tx *transaction.Transaction, group entity.FilterGroup) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`INSERT INTO filter_group (
			group_id,
			filter
		) VALUES ($1, $2)`,
		group.Group.ID,
		group.Filter,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (f *FilterGroup) UpdateFilterGroup(ct context.Context, tx *transaction.Transaction, group entity.FilterGroup) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`UPDATE filter_group
			SET filter = $1
			WHERE group_id = $2`,
		group.Filter,
		group.Group.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (f *FilterGroup) DeleteFilterGroup(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM filter_group
		WHERE group_id = $1`,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewFilterGroup(transactionFactory transaction.Factory) *FilterGroup {
	return &FilterGroup{
		transactionFactory: transactionFactory,
	}
}
