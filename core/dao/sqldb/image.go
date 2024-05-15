package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Image struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.Image = (*Image)(nil)

func (i *Image) FindImagesByAttachmentListIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) ([]entity.Image, *errs.Error) {
	var images []entity.Image

	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			id,
			url,
			size,
			attachment_list_id,
			created_at,
			updated_at
		FROM image
		WHERE attachment_list_id = $1
		`,
		attachmentListID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var image entity.Image
		err := rows.Scan(
			&image.ID,
			&image.URL,
			&image.Size,
			&image.AttachmentListID,
			&image.CreatedAt,
			&image.UpdatedAt,
		)

		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		images = append(images, image)
	}

	return images, nil
}

func (i *Image) CreateImage(ct context.Context, tx *transaction.Transaction, image entity.Image) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO image (id, url, size, attachment_list_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		`,
		image.ID,
		image.URL,
		image.Size,
		image.AttachmentListID,
		image.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewImage(metrics dao.Metrics, transactionFactory transaction.Factory) *Image {
	return &Image{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
