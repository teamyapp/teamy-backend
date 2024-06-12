package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CreateUserInput struct {
	LastName   string
	FirstName  string
	ProfileURL *string
}

type UpdateUserInput struct {
	LastName  string
	FirstName string
}

type User struct {
	logger                     telemetry.Logger
	transactionGroupFactory    transaction.GroupFactory
	toggles                    feature.Toggles
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *client.Registry
	authorizer                 client.Authorizer
	stateSyncer                *realtime.StateSyncer
	transactionFactory         cloudTransaction.Factory
	cache                      *cache.TimeBasedCache[string, any]
	userDao                    dao.User
	teamMemberDao              dao.TeamMember
	userFileUploadSessionDao   dao.UserFileUploadSession
}

func (u User) Me(ct context.Context) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.toggles.EnableAuthorization {
		query := authorization.NewReadInUserQuery(currUserID, currUserID)
		hasPermission, err := u.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.User{}, err
		}

		if !hasPermission {
			return entity.User{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	if u.toggles.EnableCache {
		value, cacheErr := u.cache.Get(ct, meCacheKey(currUserID))
		if cacheErr == nil {
			return value.(entity.User), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return entity.User{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	user, err := u.userDao.FindUserByID(ct, currUserID)
	if err != nil {
		return entity.User{}, err
	}

	if u.toggles.EnableCache {
		cacheErr := u.cache.SetIfExpired(ct, meCacheKey(currUserID), user)
		if cacheErr != nil {
			return entity.User{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return user, nil
}

func (u User) FindUsersByIDs(ct context.Context, userIDs []uint64) ([]entity.User, *errs.Error) {
	var users []entity.User
	err := u.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			users, err = u.userDao.FindUsersByIDsWithTx(ct, tx, userIDs)
			return err
		})

	return users, err
}

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.toggles.EnableAuthorization {
		query := authorization.NewReadInUserQuery(currUserID, userID)
		hasPermission, err := u.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.User{}, err
		}

		if !hasPermission {
			return entity.User{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	if u.toggles.EnableCache {
		value, cacheErr := u.cache.Get(ct, findUserCacheKey(userID))
		if cacheErr == nil {
			return value.(entity.User), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return entity.User{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	user, err := u.userDao.FindUserByID(ct, userID)
	if err != nil {
		return entity.User{}, err
	}

	if u.toggles.EnableCache {
		cacheErr := u.cache.SetIfExpired(ct, findUserCacheKey(userID), user)
		if cacheErr != nil {
			return entity.User{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return user, nil
}

func (u User) CreateUser(ct context.Context, input CreateUserInput) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	user := entity.User{
		ID:         currUserID,
		CreatedAt:  time.Now().UTC(),
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		ProfileURL: input.ProfileURL,
	}
	err := u.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			return u.userDao.CreateUser(ct, tx, user)
		})
	return user, err
}

func (u User) UpdateUser(ct context.Context, userID uint64, input UpdateUserInput) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.toggles.EnableAuthorization {
		query := authorization.NewUpdateInUserQuery(currUserID, userID)
		hasPermission, err := u.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.User{}, err
		}

		if !hasPermission {
			return entity.User{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var user entity.User
	err := u.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			user, err = u.userDao.FindUserByIDWithTx(ct, tx, userID)
			if err != nil {
				return err
			}

			user.FirstName = input.FirstName
			user.LastName = input.LastName
			updatedAt := time.Now().UTC()
			user.UpdatedAt = &updatedAt
			userMutation := mutation.NewUpdateUser(
				u.logger,
				u.stateSyncer,
				u.userDao,
				u.teamMemberDao,
				user)
			rtTx.AppendMutation(userMutation)
			return userMutation.Execute(ct, tx)
		})

	return user, err
}

func (u User) CreateUserProfileUploadSession(ct context.Context) (uint64, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return 0, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.toggles.EnableAuthorization {
		query := authorization.NewUpdateInUserQuery(currUserID, currUserID)
		hasPermission, err := u.authorizer.HasPermission(ct, query)
		if err != nil {
			return 0, err
		}

		if !hasPermission {
			return 0, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	res, rpcErr := u.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return 0, internalErr
	}

	fileUploadSession := entity.UserFileUploadSession{
		UserID:              currUserID,
		Type:                entity.ProfileUserFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err := u.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			return u.userFileUploadSessionDao.CreateUserFileUploadSession(ct, tx, fileUploadSession)
		})

	return res.UploadSessionId, err
}

func (u User) FinishUserProfileUploadSession(ct context.Context, fileUploadSessionID uint64) (entity.User, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.toggles.EnableAuthorization {
		query := authorization.NewUpdateInUserQuery(userID, userID)
		hasPermission, err := u.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.User{}, err
		}

		if !hasPermission {
			return entity.User{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	findUploadSessionReq := pbcloud.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := u.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.User{}, internalErr
	}

	var user entity.User
	err := u.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			profileUploadSession, err := u.userFileUploadSessionDao.FindUserFileUploadSessionByUserIDWithTx(
				ct,
				tx,
				userID,
				entity.ProfileUserFileUploadSessionType,
				fileUploadSessionID)
			if err != nil {
				return err
			}

			if profileUploadSession.IsCompleted {
				return errs.NewError(errs.InvalidOperation, fmt.Sprintf("profile upload session is already completed: userID=%v, fileUploadSessionID=%v",
					userID,
					fileUploadSessionID))
			}

			now := time.Now().UTC()
			profileUploadSession.IsCompleted = true
			profileUploadSession.UpdatedAt = &now
			err = u.userFileUploadSessionDao.UpdateUserFileUploadSession(ct, tx, profileUploadSession)
			if err != nil {
				return err
			}

			user, err = u.userDao.FindUserByIDWithTx(ct, tx, userID)
			if err != nil {
				return err
			}

			profileURL := io.GetFileURL(u.cloudWebAPIExternalBaseURL, uploadSession.FileId)
			user.ProfileURL = &profileURL
			user.UpdatedAt = &now
			err = u.userDao.UpdateUser(ct, tx, user)
			return err
		})

	return user, err
}

func NewUser(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	toggles feature.Toggles,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	stateSyncer *realtime.StateSyncer,
	transactionFactory cloudTransaction.Factory,
	cache *cache.TimeBasedCache[string, any],
	userDao dao.User,
	teamMemberDao dao.TeamMember,
	userFileUploadSessionDao dao.UserFileUploadSession,
) User {
	return User{
		logger:                     logger,
		transactionGroupFactory:    transactionGroupFactory,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		toggles:                    toggles,
		authorizer:                 authorizer,
		stateSyncer:                stateSyncer,
		transactionFactory:         transactionFactory,
		cache:                      cache,
		userDao:                    userDao,
		teamMemberDao:              teamMemberDao,
		userFileUploadSessionDao:   userFileUploadSessionDao,
	}
}

func findUserCacheKey(userID uint64) string {
	return fmt.Sprintf("FindUserByID(%v)", userID)
}

func meCacheKey(userID uint64) string {
	return fmt.Sprintf("Me(%v)", userID)
}
