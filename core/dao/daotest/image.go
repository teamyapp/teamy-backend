package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Image struct {
	db *dbtest.InMemoryDB
}

var _ dao.Image = (*Image)(nil)

func (i *Image) FindImagesByAttachmentListIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) ([]entity.Image, *errs.Error) {
	var images []entity.Image
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(ImageTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currImage := rawRow.(entity.Image)
				if currImage.AttachmentListID == attachmentListID {
					images = append(images, currImage)
				}
			}

			return nil
		},
	})

	return images, err
}

func (i *Image) CreateImage(ct context.Context, tx *transaction.Transaction, image entity.Image) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(ImageTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currImage := rawRow.(entity.Image)
				if currImage.AttachmentListID == image.AttachmentListID && currImage.URL == image.URL {
					return errs.NewError(errs.AlreadyExists, "Image already exists")
				}
			}

			table.Rows = append(table.Rows, image)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := i.db.GetTable(ImageTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currImage := rawRow.(entity.Image)
				if currImage.AttachmentListID == image.AttachmentListID && currImage.URL == image.URL {
					table.Rows = append(table.Rows[:i], table.Rows[i+1:]...)
					return nil
				}
			}

			return errs.NewError(errs.NotFound, "Image not found")
		},
	})
}

func NewImage(db *dbtest.InMemoryDB) *Image {
	return &Image{db: db}
}
