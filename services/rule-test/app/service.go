package app

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/EbumbaE/bandit/pkg/genproto/rule-test/api"
)

type TestProvider interface {
	DoEfficiencyTest(ctx context.Context, cycleAmount int) error
	DoLoadTest(ctx context.Context, parallelCount, cycleAmount int) error
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

	parallelCount, cycleAmount := int(req.GetParallelCount()), int(req.GetCycleAmount())

	if parallelCount <= 0 || cycleAmount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "parallelCount or cycleAmount is invalid")
	}

	if err := i.testProvider.DoLoadTest(ctx, parallelCount, cycleAmount); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (i *Implementation) DoEfficiencyTest(ctx context.Context, req *desc.EfficiencyTestRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/DoEfficiencyTest")
	defer span.Finish()

	cycleAmount := int(req.GetCycleAmount())

	if cycleAmount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "cycleAmount is invalid")
	}

	if err := i.testProvider.DoEfficiencyTest(ctx, cycleAmount); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}
