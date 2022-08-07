package service

import (
	"context"
	"fmt"
	"log"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/emptypb"
)

type User struct {
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	userDao                    dao.User
	userFileUploadSessionDao   dao.UserFileUploadSession
}

func (u User) CreateUserProfileUploadSession(ct context.Context) (uint64, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	res, err := u.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if err != nil {
		log.Println(err)
		return 0, err
	}

	fileUploadSession := entity.UserFileUploadSession{
		UserID:              userID,
		Type:                entity.ProfileUserFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err = u.userFileUploadSessionDao.CreateUserFileUploadSession(fileUploadSession)
	return res.UploadSessionId, err
}

func (u User) FinishUserProfileUploadSession(ct context.Context, fileUploadSessionID uint64) (entity.User, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return entity.User{}, err
	}

	profileUploadSession, err := u.userFileUploadSessionDao.FindUserFileUploadSessionByUserID(
		userID,
		entity.ProfileUserFileUploadSessionType,
		fileUploadSessionID)
	if err != nil {
		log.Println(err)
		return entity.User{}, err
	}

	if profileUploadSession.IsCompleted {
		err = fmt.Errorf("profile upload session is already completed: userID=%v, fileUploadSessionID=%v",
			userID, fileUploadSessionID)
		log.Println(err)
		return entity.User{}, err
	}

	now := time.Now()
	profileUploadSession.IsCompleted = true
	profileUploadSession.UpdatedAt = &now
	err = u.userFileUploadSessionDao.UpdateUserFileUploadSession(profileUploadSession)
	if err != nil {
		log.Println(err)
		return entity.User{}, err
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}

	uploadSession, err := u.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if err != nil {
		log.Println(err)
		return entity.User{}, err
	}

	user, err := u.userDao.FindUserByID(userID)
	if err != nil {
		log.Println(err)
		return entity.User{}, err
	}

	profileURL := io.GetFileURL(u.cloudWebAPIExternalBaseURL, uploadSession.FileId)
	user.ProfileURL = &profileURL
	user.UpdatedAt = &now
	err = u.userDao.UpdateUser(user)
	return user, err
}

func NewUser(
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	userDao dao.User,
	userFileUploadSessionDao dao.UserFileUploadSession,
) User {
	return User{
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		userDao:                    userDao,
		userFileUploadSessionDao:   userFileUploadSessionDao,
	}
}
