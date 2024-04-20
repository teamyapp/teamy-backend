package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const appTagRelationDaoName = "AppTagRelation"

type AppTagRelation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.AppTagRelation = (*AppTagRelation)(nil)

func (a *AppTagRelation) FindAppIDsByTagValuesWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	tagValues []string,
) ([]uint64, *errs.Error) {
	a.metrics.ReportDaoOperation(appTagRelationDaoName, "FindAppIDsByTagValuesWithTx")
	tagsStr := strings.Join(tagValues, ",")
	query := fmt.Sprintf(
		`SELECT
			app_id
		FROM app_tag_relation
		WHERE value IN (%s);`,
		tagsStr,
	)

	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	appIDs := []uint64{}
	for rows.Next() {
		var appID uint64
		err := rows.Scan(
			&appID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appIDs = append(appIDs, appID)
	}

	return appIDs, nil
}

func (a *AppTagRelation) FindTagIDsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]uint64, *errs.Error) {
	a.metrics.ReportDaoOperation(appTagRelationDaoName, "FindTagIDsByAppIDWithTx")
	rows, err := tx.SQLTx().QueryContext(ct,
		`SELECT
			tag_id
		FROM app_tag_relation
		WHERE app_id = $1;`,
		appID,
	)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	tagIDs := []uint64{}
	for rows.Next() {
		var tagID uint64
		err := rows.Scan(
			&tagID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		tagIDs = append(tagIDs, tagID)
	}

	return tagIDs, nil
}

func (a *AppTagRelation) FindAppTagByAppIDAndTagIDRelationWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	tagID uint64,
) (entity.AppTagRelation, *errs.Error) {
	a.metrics.ReportDaoOperation(appTagRelationDaoName, "FindAppTagByAppIDAndTagIDRelationWithTx")
	appTagRelation := entity.AppTagRelation{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT
			app_id,
			tag_id
		FROM app_tag_relation
		WHERE app_id = $1 AND tag_id = $2;`,
		appID,
		tagID,
	).Scan(
		&appTagRelation.AppID,
		&appTagRelation.TagID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AppTagRelation{},
				errs.NewError(errs.NotFound, fmt.Sprintf("appTagRelation not found: appID=%v, tagID=%v", appID, tagID))
		}

		return entity.AppTagRelation{}, errs.NewError(errs.Unknown, err.Error())
	}

	return appTagRelation, nil
}

func (a *AppTagRelation) CreateAppTagRelation(ct context.Context, tx *transaction.Transaction, appTagRelation entity.AppTagRelation) *errs.Error {
	a.metrics.ReportDaoOperation(appTagRelationDaoName, "CreateAppTagRelation")
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO app_tag_relation (app_id, tag_id)
		VALUES ($1, $2);`,
		appTagRelation.AppID,
		appTagRelation.TagID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppTagRelation) DeleteAppTagRelationByAppIDAndTagID(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	tagID uint64,
) *errs.Error {
	a.metrics.ReportDaoOperation(appTagRelationDaoName, "DeleteAppTagRelationByAppIDAndTagID")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM app_tag_relation
		WHERE app_id = $1 AND tag_id = $2;`,
		appID,
		tagID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppTagRelation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *AppTagRelation {
	return &AppTagRelation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
