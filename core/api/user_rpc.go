package api

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	pbmessage "github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
)

type UserRPC struct {
	userService service.User
	pbteamy.UnimplementedUserServiceServer
}

var _ runner.Service = (*UserRPC)(nil)
var _ pbteamy.UserServiceServer = (*UserRPC)(nil)

func (u UserRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterUserServiceServer(server, u)
	})
	return nil
}

func (u UserRPC) ListUsers(ct context.Context, listUsersRequest *pbteamy.ListUsersRequest) (*pbteamy.ListUsersResponse, error) {
	userIDs := make([]uint64, 0, len(listUsersRequest.UserIds))

	users, err := u.userService.FindUsersByIDs(ct, userIDs)
	if err != nil {
		return nil, errs.ToGRPCErr(err)
	}

	pbUsers := make([]*pbmessage.User, 0, len(users))
	for _, user := range users {
		pbUsers = append(pbUsers, &pbmessage.User{
			Id:         user.ID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			ProfileUrl: user.ProfileURL,
			CreatedAt:  toProtoTimePtr(&user.CreatedAt),
			UpdatedAt:  toProtoTimePtr(user.UpdatedAt),
		})
	}

	return &pbteamy.ListUsersResponse{
		Users: pbUsers,
	}, nil
}

func NewUserRPC(userService service.User) UserRPC {
	return UserRPC{userService: userService}
}
