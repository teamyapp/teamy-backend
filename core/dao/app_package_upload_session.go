package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppPackageUploadSession interface {
	FindAppPackageUploadSessionWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
		versionNumber int,
		fileUploadSessionID uint64,
	) (entity.AppPackageUploadSession, *errs.Error)
	CreateAppPackageUploadSession(
		ct context.Context,
		tx *transaction.Transaction,
		session entity.AppPackageUploadSession,
	) *errs.Error
	UpdateAppPackageFileUploadSession(
		ct context.Context,
		tx *transaction.Transaction,
		session entity.AppPackageUploadSession,
	) *errs.Error
}
