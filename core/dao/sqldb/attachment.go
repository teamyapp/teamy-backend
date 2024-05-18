package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Attachment struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.Attachment = (*Attachment)(nil)

func (i *Attachment) FindAttachmentsByAttachmentListIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) ([]entity.Attachment, *errs.Error) {
	var attachments []entity.Attachment

	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			id,
			type,
			url,
			size,
			attachment_list_id,
			created_at,
			updated_at
		FROM attachment
		WHERE attachment_list_id = $1
		`,
		attachmentListID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var attachment entity.Attachment
		err := rows.Scan(
			&attachment.ID,
			&attachment.Type,
			&attachment.URL,
			&attachment.Size,
			&attachment.AttachmentListID,
			&attachment.CreatedAt,
			&attachment.UpdatedAt,
		)

		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (i *Attachment) CreateAttachment(ct context.Context, tx *transaction.Transaction, attachment entity.Attachment) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO attachment (id, type, url, size, attachment_list_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		attachment.ID,
		attachment.Type,
		attachment.URL,
		attachment.Size,
		attachment.AttachmentListID,
		attachment.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAttachment(metrics dao.Metrics, transactionFactory transaction.Factory) *Attachment {
	return &Attachment{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
