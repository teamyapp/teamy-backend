package api

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	pbmessage "github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TeamMemberGroupRPC struct {
	teamService service.Team
	pbteamy.UnimplementedTeamMemberGroupServiceServer
}

var _ runner.Service = (*TeamMemberGroupRPC)(nil)
var _ pbteamy.TeamMemberGroupServiceServer = (*TeamMemberGroupRPC)(nil)

func (t TeamMemberGroupRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterTeamMemberGroupServiceServer(server, t)
	})
	return nil
}

func (t TeamMemberGroupRPC) ListMemberGroups(ctx context.Context, request *pbteamy.ListMemberGroupsRequest) (*pbteamy.ListTeamMemberGroupsResponse, error) {
	teamMemberGroups, err := t.teamService.FindTeamMemberGroups(ctx, request.TeamId)
	if err != nil {
		return nil, errs.ToGRPCErr(err)
	}

	groups := make([]*pbmessage.TeamMemberGroup, 0)
	for _, group := range teamMemberGroups {
		groups = append(groups, &pbmessage.TeamMemberGroup{
			Id:                       group.ID,
			Name:                     group.Name,
			OrderIndex:               int32(group.OrderIndex),
			TeamId:                   group.TeamID,
			CreatedAt:                toProtoTimePtr(&group.CreatedAt),
			UpdatedAt:                toProtoTimePtr(group.UpdatedAt),
			AuthorizationUserGroupId: group.AuthorizationUserGroupID,
		})
	}

	return &pbteamy.ListTeamMemberGroupsResponse{
		Groups: groups,
	}, nil
}

func (t TeamMemberGroupRPC) ListGroupMembers(
	ctx context.Context,
	request *pbteamy.ListGroupMembersRequest,
) (*pbteamy.ListGroupMembersResponse, error) {
	memberUsers, err := t.teamService.FindTeamMmebersByGroupID(ctx, request.GroupId)
	if err != nil {
		return nil, errs.ToGRPCErr(err)
	}

	users := make([]*pbmessage.User, 0)
	for _, memberUser := range memberUsers {
		users = append(users, &pbmessage.User{
			Id:         memberUser.ID,
			FirstName:  memberUser.FirstName,
			LastName:   memberUser.LastName,
			ProfileUrl: memberUser.ProfileURL,
			CreatedAt:  toProtoTimePtr(&memberUser.CreatedAt),
			UpdatedAt:  toProtoTimePtr(memberUser.UpdatedAt),
		})
	}

	return &pbteamy.ListGroupMembersResponse{
		Users: users,
	}, nil
}

func (t TeamMemberGroupRPC) CreateMemberGroup(ctx context.Context, request *pbteamy.CreateTeamMemberGroupRequest) (*pbteamy.CreateTeamMemberGroupResponse, error) {
	teamMemberGroup, err := t.teamService.CreateTeamMemberGroup(ctx, service.CreateTeamMemberGroupInput{
		Name:   request.Name,
		TeamID: request.TeamId,
	})
	return &pbteamy.CreateTeamMemberGroupResponse{
		GroupId: teamMemberGroup.ID,
	}, errs.ToGRPCErr(err)
}

func (t TeamMemberGroupRPC) UpdateMemberGroup(ctx context.Context, request *pbteamy.UpdateTeamMemberGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.UpdateTeamMemberGroup(ctx, service.UpdateTeamMemberGroupInput{
		GroupID: request.GroupId,
		Name:    request.Name,
	})
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamMemberGroupRPC) DeleteMemberGroup(ctx context.Context, request *pbteamy.DeleteTeamMemberGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.DeleteTeamMemberGroup(ctx, request.GroupId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamMemberGroupRPC) AddMemberToGroup(ctx context.Context, request *pbteamy.AddTeamMemberToGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.AddUserToTeamMemberGroup(ctx, request.GroupId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamMemberGroupRPC) RemoveMemberFromGroup(ctx context.Context, request *pbteamy.RemoveTeamMemberFromGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.RemoveUserFromTeamMemberGroup(ctx, request.GroupId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func NewTeamMemberGroupRPC(
	teamService service.Team,
) TeamMemberGroupRPC {
	return TeamMemberGroupRPC{
		teamService: teamService,
	}
}
