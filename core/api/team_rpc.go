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

type TeamRPC struct {
	teamService service.Team
	pbteamy.UnimplementedTeamServiceServer
}

var _ runner.Service = (*TeamRPC)(nil)
var _ pbteamy.TeamServiceServer = (*TeamRPC)(nil)

func (t TeamRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterTeamServiceServer(server, t)
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
		Name:        *request.Name,
		OwnerUserID: *request.OwnerUserId,
	})
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (t TeamRPC) ListTeamMembers(ctx context.Context, request *pbteamy.ListTeamMembersRequest) (*pbteamy.ListTeamMembersResponse, error) {
	teamMembers, internalErr := t.teamService.FindTeamMembers(ctx, request.TeamId)
	if internalErr != nil {
		return nil, errs.ToGRPCErr(internalErr)
	}

	members := make([]*pbmessage.TeamMember, 0)
	for _, member := range teamMembers {
		members = append(members, &pbmessage.TeamMember{
			UserId:          member.UserID,
			TeamId:          member.TeamID,
			WeeklyBandwidth: toProtoDurationPtr(&member.WeeklyBandwidth),
			CreatedAt:       toProtoTimePtr(&member.CreatedAt),
			UpdatedAt:       toProtoTimePtr(member.UpdatedAt),
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

func (t TeamRPC) RemoveMemberFromTeam(ctx context.Context, request *pbteamy.RemoveMemberFromTeamRequest) (*emptypb.Empty, error) {
	_, err := t.teamService.RemoveMemberFromTeam(ctx, request.TeamId, request.MemberUserId)
	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func NewTeamRPC(
	teamService service.Team,
) TeamRPC {
	return TeamRPC{
		teamService: teamService,
	}
}
