package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
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
	dataCollector              telemetry.DataCollector
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	stateSyncer                *realtime.StateSyncer
	userDao                    dao.User
	userFileUploadSessionDao   dao.UserFileUploadSession
	teamMemberDao              dao.TeamMember
}

func (u User) Me(ct context.Context) (entity.User, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.User{}, internalErr
	}

	return u.userDao.FindUserByID(ct, userID)
}

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	return u.userDao.FindUserByID(ct, userID)
}

func (u User) CreateUser(ct context.Context, input CreateUserInput) (entity.User, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.User{}, internalErr
	}

	user := entity.User{
		ID:         userID,
		CreatedAt:  time.Now(),
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		ProfileURL: input.ProfileURL,
	}

	err := u.userDao.CreateUser(ct, user)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	return user, nil
}

func (u User) UpdateUser(ct context.Context, userID uint64, input UpdateUserInput) (entity.User, *errs.Error) {
	user, err := u.userDao.FindUserByID(ct, userID)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName
	updatedAt := time.Now()
	user.UpdatedAt = &updatedAt
	realTimeTransaction := realtime.NewTransaction(u.dataCollector, u.stateSyncer)
	userMutation := mutation.NewUpdateUserMutation(
		u.dataCollector,
		u.stateSyncer,
		u.teamMemberDao,
		u.userDao,
		user)
	err = realTimeTransaction.ApplyMutation(ct, userMutation)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	return user, nil
}

func (u User) CreateUserProfileUploadSession(ct context.Context) (uint64, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	res, rpcErr := u.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	fileUploadSession := entity.UserFileUploadSession{
		UserID:              userID,
		Type:                entity.ProfileUserFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err := u.userFileUploadSessionDao.CreateUserFileUploadSession(ct, fileUploadSession)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return 0, err
	}

	return res.UploadSessionId, err
}

func (u User) FinishUserProfileUploadSession(ct context.Context, fileUploadSessionID uint64) (entity.User, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.User{}, internalErr
	}

	profileUploadSession, err := u.userFileUploadSessionDao.FindUserFileUploadSessionByUserID(
		ct,
		userID,
		entity.ProfileUserFileUploadSessionType,
		fileUploadSessionID)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	if profileUploadSession.IsCompleted {
		internalErr := &errs.Error{
			Code: errs.InvalidOperation,
			Message: fmt.Sprintf("profile upload session is already completed: userID=%v, fileUploadSessionID=%v",
				userID,
				fileUploadSessionID),
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return entity.User{}, internalErr
	}

	now := time.Now()
	profileUploadSession.IsCompleted = true
	profileUploadSession.UpdatedAt = &now
	err = u.userFileUploadSessionDao.UpdateUserFileUploadSession(ct, profileUploadSession)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}

	uploadSession, rpcErr := u.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.User{}, internalErr
	}

	user, err := u.userDao.FindUserByID(ct, userID)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.User{}, err
	}

	profileURL := io.GetFileURL(u.cloudWebAPIExternalBaseURL, uploadSession.FileId)
	user.ProfileURL = &profileURL
	user.UpdatedAt = &now
	err = u.userDao.UpdateUser(ct, user)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
	}

	return user, nil
}

func NewUser(
	dataCollector telemetry.DataCollector,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	userDao dao.User,
	userFileUploadSessionDao dao.UserFileUploadSession,
	teamMemberDao dao.TeamMember,
) User {
	return User{
		dataCollector:              dataCollector,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		stateSyncer:                stateSyncer,
		userDao:                    userDao,
		userFileUploadSessionDao:   userFileUploadSessionDao,
		teamMemberDao:              teamMemberDao,
	}
}
