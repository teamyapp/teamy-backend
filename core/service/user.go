package service

import (
	"context"
	"errors"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/emptypb"
)

type User struct {
	dataCollector              obs.DataCollector
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	userDao                    dao.User
	userFileUploadSessionDao   dao.UserFileUploadSession
}

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, error) {
	return u.userDao.FindUserByID(userID)
}

func (u User) CreateUserProfileUploadSession(ct context.Context) (uint64, error) {
	userID, err := ctx.UserIDFromContext(u.dataCollector, ct)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	res, err := u.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	return res.UploadSessionId, err
}

func (u User) FinishUserProfileUploadSession(ct context.Context, fileUploadSessionID uint64) (entity.User, error) {
	userID, err := ctx.UserIDFromContext(u.dataCollector, ct)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.User{}, err
	}

	profileUploadSession, err := u.userFileUploadSessionDao.FindUserFileUploadSessionByUserID(
		userID,
		entity.ProfileUserFileUploadSessionType,
		fileUploadSessionID)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.User{}, err
	}

	if profileUploadSession.IsCompleted {
		err = errors.New("profile upload session is already completed")
		u.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"userID":              userID,
				"fileUploadSessionID": fileUploadSessionID,
			},
		})
		return entity.User{}, err
	}

	now := time.Now()
	profileUploadSession.IsCompleted = true
	profileUploadSession.UpdatedAt = &now
	err = u.userFileUploadSessionDao.UpdateUserFileUploadSession(profileUploadSession)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.User{}, err
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}

	uploadSession, err := u.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.User{}, err
	}

	user, err := u.userDao.FindUserByID(userID)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.User{}, err
	}

	profileURL := io.GetFileURL(u.cloudWebAPIExternalBaseURL, uploadSession.FileId)
	user.ProfileURL = &profileURL
	user.UpdatedAt = &now
	err = u.userDao.UpdateUser(user)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return user, nil
}

func NewUser(
	dataCollector obs.DataCollector,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	userDao dao.User,
	userFileUploadSessionDao dao.UserFileUploadSession,
) User {
	return User{
		dataCollector:              dataCollector,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		userDao:                    userDao,
		userFileUploadSessionDao:   userFileUploadSessionDao,
	}
}
