package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AttachmentList struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.AttachmentList = (*AttachmentList)(nil)

func (a *AttachmentList) FindAttachmentListByIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) (entity.AttachmentList, *errs.Error) {
	var attachmentList entity.AttachmentList
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			owner_type,
			owner_id,
			list_label,
			list_id,
			created_at,
			updated_at
		FROM attachment_list
		WHERE list_id = $1
		`,
		attachmentListID,
	).Scan(
		&attachmentList.OwnerType,
		&attachmentList.OwnerID,
		&attachmentList.ListLabel,
		&attachmentList.ListID,
		&attachmentList.CreatedAt,
		&attachmentList.UpdatedAt,
	)

	if err != nil {
		return entity.AttachmentList{}, errs.NewError(errs.Unknown, err.Error())
	}

	return attachmentList, nil
}

func (a *AttachmentList) FindAttachmentListsByOwnerIDAndOwnerTypeWithTx(
    ct context.Context, 
    tx *transaction.Transaction, 
    ownerType entity.AttachmentListOwnerType,
    ownerID uint64,
) ([]entity.AttachmentList, *errs.Error) {
	var attachmentLists []entity.AttachmentList
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			owner_type,
			owner_id,
			list_label,
			list_id,
			created_at,
			updated_at
		FROM attachment_list
		WHERE owner_id = $1 AND owner_type = $2
		`,
		ownerID,
		ownerType,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var attachmentList entity.AttachmentList
		err := rows.Scan(
			&attachmentList.OwnerType,
			&attachmentList.OwnerID,
			&attachmentList.ListLabel,
			&attachmentList.ListID,
			&attachmentList.CreatedAt,
			&attachmentList.UpdatedAt,
		)

		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		attachmentLists = append(attachmentLists, attachmentList)
	}

	return attachmentLists, nil
}

func (a *AttachmentList) FindAttachmentListsWithTx(ct context.Context, tx *transaction.Transaction, ownerID uint64, ownerType entity.AttachmentListOwnerType, ListLabel string) ([]entity.AttachmentList, *errs.Error) {
	var attachmentLists []entity.AttachmentList
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			owner_type,
			owner_id,
			list_label,
			list_id,
			created_at,
			updated_at
		FROM attachment_list
		WHERE owner_id = $1 AND owner_type = $2 AND list_label = $3
		`,
		ownerID,
		ownerType,
		ListLabel,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var attachmentList entity.AttachmentList
		err := rows.Scan(
			&attachmentList.OwnerType,
			&attachmentList.OwnerID,
			&attachmentList.ListLabel,
			&attachmentList.ListID,
			&attachmentList.CreatedAt,
			&attachmentList.UpdatedAt,
		)

		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		attachmentLists = append(attachmentLists, attachmentList)
	}

	return attachmentLists, nil
}

func (a *AttachmentList) CreateAttachmentList(ct context.Context, tx *transaction.Transaction, attachmentList entity.AttachmentList) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO attachment_list (owner_type, owner_id, list_label, list_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		`,
		attachmentList.OwnerType,
		attachmentList.OwnerID,
		attachmentList.ListLabel,
		attachmentList.ListID,
		attachmentList.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAttachmentList(metrics dao.Metrics, transactionFactory transaction.Factory) *AttachmentList {
	return &AttachmentList{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
