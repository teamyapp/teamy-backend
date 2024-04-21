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

const tagDaoName = "Tag"

type Tag struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.Tag = (*Tag)(nil)

func (t *Tag) FindTagsByTagIDsWithTx(ct context.Context, tx *transaction.Transaction, tagIDs []uint64) ([]entity.Tag, *errs.Error) {
	t.metrics.ReportDaoOperation(tagDaoName, "FindTagsByTagIDsWithTx")
	idsString := toIDsString(tagIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		value
	FROM tag
	WHERE id IN (%s);`,
		idsString)
	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var tags []entity.Tag
	for rows.Next() {
		tag := entity.Tag{}
		err := rows.Scan(
			&tag.ID,
			&tag.Value,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (t *Tag) FindTagByValueWithTx(ct context.Context, tx *transaction.Transaction, value string) (entity.Tag, *errs.Error) {
	t.metrics.ReportDaoOperation(tagDaoName, "FindTagByValueWithTx")
	tag := entity.Tag{}
	err := tx.SQLTx().QueryRowContext(ct, `
	SELECT
		id,
		value
	FROM tag
	WHERE value = $1;`,
		value).
		Scan(
			&tag.ID,
			&tag.Value,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Tag{}, errs.NewError(errs.NotFound, fmt.Sprintf("tag not found: value=%v", value))
		}

		return entity.Tag{}, errs.NewError(errs.Unknown, err.Error())
	}

	return tag, nil
}

func (t *Tag) FindTagByIDWithTx(ct context.Context, tx *transaction.Transaction, tagID uint64) (entity.Tag, *errs.Error) {
	t.metrics.ReportDaoOperation(tagDaoName, "FindTagByIDWithTx")
	tag := entity.Tag{}
	err := tx.SQLTx().QueryRowContext(ct, `
	SELECT
		id,
		value
	FROM tag
	WHERE id = $1;`,
		tagID).
		Scan(
			&tag.ID,
			&tag.Value,
		)
	if err != nil {
		return entity.Tag{}, errs.NewError(errs.Unknown, err.Error())
	}

	return tag, nil
}

func (t *Tag) CreateTag(ct context.Context, tx *transaction.Transaction, tag entity.Tag) *errs.Error {
	t.metrics.ReportDaoOperation(tagDaoName, "CreateTag")
	_, err := tx.SQLTx().ExecContext(ct, `
	INSERT INTO tag (
		id,
		value
	) VALUES (
		$1,
		$2
	);`,
		tag.ID,
		tag.Value,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t *Tag) DeleteTag(ct context.Context, tx *transaction.Transaction, tagID uint64) *errs.Error {
	t.metrics.ReportDaoOperation(tagDaoName, "DeleteTag")
	_, err := tx.SQLTx().ExecContext(ct, `
	DELETE FROM tag
	WHERE id = $1;`,
		tagID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTag(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *Tag {
	return &Tag{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
