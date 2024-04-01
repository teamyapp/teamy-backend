package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Project interface {
	FindProjectByIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) (entity.Project, *errs.Error)
	FindProjectsByIDsWithTx(ct context.Context, tx *transaction.Transaction, projectIDs []uint64) ([]entity.Project, *errs.Error)
	FindProjectsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Project, *errs.Error)
	CreateProject(ct context.Context, tx *transaction.Transaction, project entity.Project) *errs.Error
	UpdateProject(ct context.Context, tx *transaction.Transaction, project entity.Project) *errs.Error
	DeleteProject(ct context.Context, tx *transaction.Transaction, projectID uint64) *errs.Error
}
