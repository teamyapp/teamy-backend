package api

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TeamRPC struct {
	teamService service.Team
	userService service.User
	pbteamy.UnimplementedTeamServer
}

var _ runner.Service = (*TeamRPC)(nil)
var _ pbteamy.TeamServer = (*TeamRPC)(nil)

func (t TeamRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterTeamServer(server, t)
	})
	return nil
}

func (t TeamRPC) CreateTeam(ctx context.Context, request *pbteamy.CreateTeamRequest) (*pbteamy.CreateTeamResponse, error) {
	team, err := t.teamService.CreateTeam(ctx, service.CreateTeamInput{
		Name: request.Name,
	})
	if err != nil {
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreateTeamResponse{
		TeamId: team.ID,
	}, nil
}

func (t TeamRPC) UpdateTeam(ctx context.Context, request *pbteamy.UpdateTeamRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.UpdateTeam(ctx, request.TeamId, service.UpdateTeamInput{
		Name:        request.Name,
		OwnerUserID: request.OwnerUserId,
	})
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) ListTeamMembers(ctx context.Context, request *pbteamy.ListTeamMembersRequest) (*pbteamy.ListTeamMembersResponse, error) {
	teamMembers, internalErr := t.teamService.FindTeamMembers(ctx, request.TeamId)
	if internalErr != nil {
		return nil, errs.ToGRPCErr(internalErr)
	}

	members := make([]*pbteamy.TeamMember, 0)
	for _, member := range teamMembers {
		user, internalErr := t.userService.FindUserByID(ctx, member.UserID)
		if internalErr != nil {
			return nil, errs.ToGRPCErr(internalErr)
		}

		members = append(members, &pbteamy.TeamMember{
			UserId:     member.UserID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			ProfileUrl: user.ProfileURL,
		})
	}

	return &pbteamy.ListTeamMembersResponse{
		TeamMembers: members,
	}, nil
}

func (t TeamRPC) AddMemberToTeam(ctx context.Context, request *pbteamy.AddMemberToTeamRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.AddMemberToTeam(ctx, request.TeamId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) RemoveMemberFromTeam(ctx context.Context, request *pbteamy.RemoveMemberFromRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.RemoveMemberFromTeam(ctx, request.TeamId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) ListMemberGroups(ctx context.Context, request *pbteamy.ListMemberGroupsRequest) (*pbteamy.ListTeamMemberGroupsResponse, error) {
	teamMemberGroups, err := t.teamService.FindTeamMemberGroups(ctx, request.TeamId)
	if err != nil {
		return nil, errs.ToGRPCErr(err)
	}

	groups := make([]*pbteamy.TeamMemberGroup, 0)
	for _, group := range teamMemberGroups {
		groups = append(groups, &pbteamy.TeamMemberGroup{
			GroupId:       group.ID,
			Name:          group.Name,
			MemberUserIds: group.MemberUserIDs,
		})
	}

	return &pbteamy.ListTeamMemberGroupsResponse{
		Groups: groups,
	}, nil
}

func (t TeamRPC) CreateMemberGroup(ctx context.Context, request *pbteamy.CreateTeamMemberGroupRequest) (*pbteamy.CreateTeamMemberGroupResponse, error) {
	teamMemberGroup, err := t.teamService.CreateTeamMemberGroup(ctx, service.CreateTeamMemberGroupInput{
		Name:   request.Name,
		TeamID: request.TeamId,
	})
	return &pbteamy.CreateTeamMemberGroupResponse{
		GroupId: teamMemberGroup.ID,
	}, errs.ToGRPCErr(err)
}

func (t TeamRPC) UpdateMemberGroup(ctx context.Context, request *pbteamy.UpdateTeamMemberGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.UpdateTeamMemberGroup(ctx, service.UpdateTeamMemberGroupInput{
		GroupID: request.GroupId,
		Name:    request.Name,
	})
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) DeleteMemberGroup(ctx context.Context, request *pbteamy.DeleteTeamMemberGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.DeleteTeamMemberGroup(ctx, request.GroupId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) AddMemberToGroup(ctx context.Context, request *pbteamy.AddTeamMemberToGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.AddUserToTeamMemberGroup(ctx, request.GroupId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) RemoveMemberFromGroup(ctx context.Context, request *pbteamy.RemoveTeamMemberFromGroupRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.RemoveUserFromTeamMemberGroup(ctx, request.GroupId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func NewTeamRPC(
	teamService service.Team,
	userService service.User,
) TeamRPC {
	return TeamRPC{
		teamService: teamService,
		userService: userService,
	}
}
