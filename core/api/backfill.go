package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BackFillRPC struct {
	dataCollector   telemetry.DataCollector
	backFillService service.Backfill
	proto.UnimplementedBackFillServer
}

var _ runner.Service = (*BackFillRPC)(nil)
var _ proto.BackFillServer = (*BackFillRPC)(nil)

func (b BackFillRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterBackFillServer(server, b)
	})
	return nil
}

func (b BackFillRPC) BackFillPullRequestLinks(ct context.Context, in *proto.BackFillPullRequestLinksRequest) (*emptypb.Empty, error) {
	panic("implement me")
	return &emptypb.Empty{}, nil
}

func (b BackFillRPC) BackFillParticipantsBandwidth(ct context.Context, in *proto.BackFillParticipantsBandwidthRequest) (*emptypb.Empty, error) {
	panic("implement me")
	return &emptypb.Empty{}, nil
}

func NewBackFillRPC(dataCollector telemetry.DataCollector, backFillService service.Backfill) BackFillRPC {
	return BackFillRPC{
		dataCollector:   dataCollector,
		backFillService: backFillService,
	}
}
