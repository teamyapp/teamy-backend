package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type Backfill struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	cloudClientRegistry     *client.Registry
	taskDao                 dao.Task
	teamDao                 dao.Team
	attachmentListDao       dao.AttachmentList
	teamMemberGroupDao      dao.TeamMemberGroup
	teamMemberGroupRepo     repository.TeamMemberGroup
}

func (b *Backfill) backfillTeamMemberGroupOrder(ct context.Context) *errs.Error {
	return b.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		teams, internalErr := b.teamDao.FindAllTeamsWithTx(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		now := time.Now().UTC()
		for _, team := range teams {
			memberGroups, internalErr := b.teamMemberGroupRepo.FindMemberGroupsByTeamID(ct, tx, team.ID)
			if internalErr != nil {
				return internalErr
			}

			for index, memberGroup := range memberGroups {
				if memberGroup.Order == index {
					continue
				}

				rawTeamMemberGroup, internalErr := b.teamMemberGroupDao.FindMemberGroupByID(ct, tx, memberGroup.ID)
				if internalErr != nil {
					return internalErr
				}

				rawTeamMemberGroup.Order = index
				rawTeamMemberGroup.UpdatedAt = &now
				internalErr = b.teamMemberGroupDao.UpdateMemberGroup(ct, tx, rawTeamMemberGroup)
				if internalErr != nil {
					return internalErr
				}
			}
		}

		return nil
	})
}

func (b *Backfill) backfillAttachmentListForTaskContext(ct context.Context) *errs.Error {
	return b.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		tasks, internalErr := b.taskDao.FindAllTasksWithTx(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		for _, task := range tasks {
			attachmentLists, internalErr := b.attachmentListDao.FindAttachmentListsWithTx(ct, tx, entity.AttachmentListOwnerTypeTask, task.ID, "context")
			if internalErr != nil {
				return internalErr
			}

			if len(attachmentLists) > 0 {
				continue
			}
			genAttachmentListIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "attachmentListID"}
			genAttachmentListIDRes, rpcErr := b.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAttachmentListIDReq)
			if rpcErr != nil {
				return errs.FromGRPCErr(rpcErr)
			}

			attachmentList := entity.AttachmentList{
				OwnerType: entity.AttachmentListOwnerTypeTask,
				OwnerID:   task.ID,
				ListLabel: "context",
				ListID:    genAttachmentListIDRes.UniqueNumber,
				CreatedAt: time.Now().UTC(),
			}

			internalErr = b.attachmentListDao.CreateAttachmentList(ct, tx, attachmentList)
			if internalErr != nil {
				return internalErr
			}
		}

		return nil
	})
}

func (b *Backfill) BackfillData(ct context.Context) *errs.Error {
	return b.backfillAttachmentListForTaskContext(ct)
}

func NewBackfill(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	cloudClientRegistry *client.Registry,
	taskDao dao.Task,
	teamDao dao.Team,
	attachmentListDao dao.AttachmentList,
	teamMemberGroupDao dao.TeamMemberGroup,
	teamMemberGroupRepo repository.TeamMemberGroup,
) *Backfill {
	return &Backfill{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		cloudClientRegistry:     cloudClientRegistry,
		taskDao:                 taskDao,
		teamDao:                 teamDao,
		attachmentListDao:       attachmentListDao,
		teamMemberGroupDao:      teamMemberGroupDao,
		teamMemberGroupRepo:     teamMemberGroupRepo,
	}
}
