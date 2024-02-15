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

type Tag struct {
	transactionFactory transaction.Factory
}

var _ dao.Tag = (*Tag)(nil)

func (*Tag) FindTagsByTagIDsWithTx(ct context.Context, tx *transaction.Transaction, tagIDs []uint64) ([]entity.Tag, *errs.Error) {
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

	tags := []entity.Tag{}
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

func (*Tag) FindTagByValueWithTx(ct context.Context, tx *transaction.Transaction, value string) (entity.Tag, *errs.Error) {
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

func (*Tag) FindTagByIDWithTx(ct context.Context, tx *transaction.Transaction, tagID uint64) (entity.Tag, *errs.Error) {
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

func (*Tag) CreateTag(ct context.Context, tx *transaction.Transaction, tag entity.Tag) *errs.Error {
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

func NewTag(transactionFactory transaction.Factory) *Tag {
	return &Tag{
		transactionFactory: transactionFactory,
	}
}
