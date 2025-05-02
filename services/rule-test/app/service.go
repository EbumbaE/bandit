package app

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/EbumbaE/bandit/pkg/genproto/rule-test/api"
)

type TestProvider interface {
	DoEfficiencyTest(ctx context.Context, targetRPS int, duration time.Duration) error
	DoLoadTest(ctx context.Context, parallelCount, targetRPS int, duration time.Duration) error
}

type Implementation struct {
	testProvider TestProvider

	desc.UnimplementedRuleTestServiceServer
}

func NewService(testProvider TestProvider) *Implementation {
	return &Implementation{
		testProvider: testProvider,
	}
}

func (i *Implementation) DoLoadTest(ctx context.Context, req *desc.LoadTestRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/DoLoadTest")
	defer span.Finish()

	parallelCount, targetRPS := int(req.GetParallelCount()), int(req.GetTargetRps())
	duration := req.GetDuration().AsDuration()

	if parallelCount <= 0 || targetRPS <= 0 || duration <= 0 {
		return nil, status.Error(codes.InvalidArgument, "parallel_count or target_rps or duration is invalid")
	}

	if err := i.testProvider.DoLoadTest(ctx, parallelCount, targetRPS, duration); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (i *Implementation) DoEfficiencyTest(ctx context.Context, req *desc.EfficiencyTestRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/DoEfficiencyTest")
	defer span.Finish()

	targetRPS, duration := int(req.GetTargetRps()), req.GetDuration().AsDuration()

	if targetRPS <= 0 || duration <= 0 {
		return nil, status.Error(codes.InvalidArgument, "target_rps or duration is invalid")
	}

	if err := i.testProvider.DoEfficiencyTest(ctx, targetRPS, duration); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}
