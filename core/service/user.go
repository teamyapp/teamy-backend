package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
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
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *client.Registry
	stateSyncer                *realtime.StateSyncer
	authorizer                 client.Authorizer
	transactionFactory         transaction.Factory
	featureToggles             feature.Toggles
	userDaoV2                  daov2.User
	teamMemberDaoV2            daov2.TeamMember
	userFileUploadSessionDaoV2 daov2.UserFileUploadSession
}

func (u User) Me(ct context.Context) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.featureToggles.EnableAuthorization {
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

	return u.userDaoV2.FindUserByID(ct, currUserID)
}

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.featureToggles.EnableAuthorization {
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

	return u.userDaoV2.FindUserByID(ct, userID)
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
	txCtx := TransactionsContext{
		logger:             u.logger,
		transactionFactory: u.transactionFactory,
		stateSyncer:        u.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return u.userDaoV2.CreateUser(ct, tx, user)
	})
	return user, err
}

func (u User) UpdateUser(ct context.Context, userID uint64, input UpdateUserInput) (entity.User, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.featureToggles.EnableAuthorization {
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
	txCtx := TransactionsContext{
		logger:             u.logger,
		transactionFactory: u.transactionFactory,
		stateSyncer:        u.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		user, err = u.userDaoV2.FindUserByIDWithTx(ct, tx, userID)
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
			u.userDaoV2,
			u.teamMemberDaoV2,
			user)
		rtTx.AppendMutation(userMutation)
		return userMutation.ExecuteV2(ct, tx)
	})

	return user, err
}

func (u User) CreateUserProfileUploadSession(ct context.Context) (uint64, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return 0, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.featureToggles.EnableAuthorization {
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

	txCtx := TransactionsContext{
		logger:             u.logger,
		transactionFactory: u.transactionFactory,
		stateSyncer:        u.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return u.userFileUploadSessionDaoV2.CreateUserFileUploadSession(ct, tx, fileUploadSession)
	})

	return res.UploadSessionId, err
}

func (u User) FinishUserProfileUploadSession(ct context.Context, fileUploadSessionID uint64) (entity.User, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.User{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if u.featureToggles.EnableAuthorization {
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

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := u.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.User{}, internalErr
	}

	var user entity.User
	txCtx := TransactionsContext{
		logger:             u.logger,
		transactionFactory: u.transactionFactory,
		stateSyncer:        u.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		profileUploadSession, err := u.userFileUploadSessionDaoV2.FindUserFileUploadSessionByUserIDWithTx(
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
		err = u.userFileUploadSessionDaoV2.UpdateUserFileUploadSession(ct, tx, profileUploadSession)
		if err != nil {
			return err
		}

		user, err = u.userDaoV2.FindUserByIDWithTx(ct, tx, userID)
		if err != nil {
			return err
		}

		profileURL := io.GetFileURL(u.cloudWebAPIExternalBaseURL, uploadSession.FileId)
		user.ProfileURL = &profileURL
		user.UpdatedAt = &now
		err = u.userDaoV2.UpdateUser(ct, tx, user)
		return err
	})

	return user, err
}

func NewUser(
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	featureToggles feature.Toggles,
	userDaoV2 daov2.User,
	teamMemberDaoV2 daov2.TeamMember,
	userFileUploadSessionDaoV2 daov2.UserFileUploadSession,
) User {
	return User{
		logger:                     logger,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		featureToggles:             featureToggles,
		authorizer:                 authorizer,
		stateSyncer:                stateSyncer,
		transactionFactory:         transactionFactory,
		userDaoV2:                  userDaoV2,
		teamMemberDaoV2:            teamMemberDaoV2,
		userFileUploadSessionDaoV2: userFileUploadSessionDaoV2,
	}
}
