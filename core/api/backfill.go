package api

import (
	"context"
	_ "embed"

	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BackfillRPC struct {
	dataCollector   telemetry.DataCollector
	backfillService service.Backfill
	proto.UnimplementedBackfillServer
}

var _ runner.Service = (*BackfillRPC)(nil)
var _ proto.BackfillServer = (*BackfillRPC)(nil)

func (b BackfillRPC) Start(runner *runner.ServiceRunner) error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterBackfillServer(server, b)
	})
	return nil
}

func (b BackfillRPC) BackfillPullRequestLinks(ct context.Context, in *proto.BackfillPullRequestLinksRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (b BackfillRPC) BackfillParticipantsBandwidth(ct context.Context, in *proto.BackfillParticipantsBandwidthRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func NewBackfillRPC(dataCollector telemetry.DataCollector, backfillService service.Backfill) BackfillRPC {
	return BackfillRPC{
		dataCollector:   dataCollector,
		backfillService: backfillService,
	}
}
