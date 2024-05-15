package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AttachmentList struct {
	db *dbtest.InMemoryDB
}

var _ dao.AttachmentList = (*AttachmentList)(nil)

func (a *AttachmentList) FindAttachmentListByIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) (entity.AttachmentList, *errs.Error) {
	var attachmentList entity.AttachmentList
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AttachmentListTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAttachmentList := rawRow.(entity.AttachmentList)
				if currAttachmentList.ListID == attachmentListID {
					attachmentList = currAttachmentList
					break
				}
			}

			return nil
		},
	})

	return attachmentList, err
}

func (a *AttachmentList) CreateAttachmentList(ct context.Context, tx *transaction.Transaction, attachmentList entity.AttachmentList) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AttachmentListTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAttachmentList := rawRow.(entity.AttachmentList)
				if currAttachmentList.OwnerID == attachmentList.OwnerID && currAttachmentList.OwnerType == attachmentList.OwnerType && currAttachmentList.ListLabel == attachmentList.ListLabel {
					return errs.NewError(errs.AlreadyExists, "Attachment list already exists")
				}
			}

			table.Rows = append(table.Rows, attachmentList)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AttachmentListTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currAttachmentList := rawRow.(entity.AttachmentList)
				if currAttachmentList.OwnerID == attachmentList.OwnerID && currAttachmentList.OwnerType == attachmentList.OwnerType && currAttachmentList.ListLabel == attachmentList.ListLabel {
					table.Rows = append(table.Rows[:i], table.Rows[i+1:]...)
					return nil
				}
			}

			return errs.NewError(errs.NotFound, "Attachment list not found")
		},
	})

}

func (a *AttachmentList) FindAttachmentListsWithTx(ct context.Context, tx *transaction.Transaction, ownerID uint64, ownerType entity.AttachmentListOwnerType, ListLabel string) ([]entity.AttachmentList, *errs.Error) {
	var attachmentLists []entity.AttachmentList
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AttachmentListTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAttachmentList := rawRow.(entity.AttachmentList)
				if currAttachmentList.OwnerID == ownerID && currAttachmentList.OwnerType == ownerType && currAttachmentList.ListLabel == ListLabel {
					attachmentLists = append(attachmentLists, currAttachmentList)
				}
			}

			return nil
		},
	})

	return attachmentLists, err
}

func (a *AttachmentList) FindAttachmentListsByOwnerIDAndOwnerTypeWithTx(ct context.Context, tx *transaction.Transaction, ownerID uint64, ownerType entity.AttachmentListOwnerType) ([]entity.AttachmentList, *errs.Error) {
	var attachmentLists []entity.AttachmentList
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AttachmentListTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAttachmentList := rawRow.(entity.AttachmentList)
				if currAttachmentList.OwnerID == ownerID && currAttachmentList.OwnerType == ownerType {
					attachmentLists = append(attachmentLists, currAttachmentList)
				}
			}

			return nil
		},
	})

	return attachmentLists, err
}

func NewAttachmentList(db *dbtest.InMemoryDB) *AttachmentList {
	return &AttachmentList{
		db: db,
	}
}
