package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/runner"

	"google.golang.org/grpc"
)

type Authorization struct {
	authorizationService service.Authorization
	proto.UnimplementedAuthorizationServer
}

var _ runner.Service = (*Authorization)(nil)
var _ proto.AuthorizationServer = (*Authorization)(nil)

func (a Authorization) HasPermission(ctx context.Context, req *proto.HasPermissionRequest) (*proto.HasPermissionResponse, error) {
	hasPermission, err := a.authorizationService.HasPermission(req.ResourceType, req.ResourceId, req.Operation, req.UserId)
	if err != nil {
		return nil, err
	}

	return &proto.HasPermissionResponse{HasPermission: hasPermission}, nil
}

func (a Authorization) Start(rn *runner.ServiceRunner) error {
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterAuthorizationServer(server, a)
	})
	return nil
}

func NewAuthorization(authorizationService service.Authorization) Authorization {
	return Authorization{
		authorizationService: authorizationService,
	}
}
