package app

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
	"github.com/EbumbaE/bandit/services/rule-diller/internal/provider"
)

type DillerProvider interface {
	GetRuleData(ctx context.Context, service, ctxKey string) (string, string, error)
	GetRuleStatistic(ctx context.Context, service, ctxKey string) ([]model.Variant, error)
}

type Implementation struct {
	dillerProvider DillerProvider

	desc.UnimplementedRuleDillerServiceServer
}

func NewService(dillerProvider DillerProvider) *Implementation {
	return &Implementation{
		dillerProvider: dillerProvider,
	}
}

func (i *Implementation) GetRuleData(ctx context.Context, req *desc.GetRuleRequest) (*desc.GetRuleDataResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetRule")
	defer span.Finish()

	if len(req.GetService()) == 0 || len(req.GetContext()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty service or context")
	}

	ruleData, payload, err := i.dillerProvider.GetRuleData(ctx, req.GetService(), req.GetContext())
	if err != nil {
		if errors.Is(err, provider.ErrEmptyAnswer) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.GetRuleDataResponse{
		Data:    ruleData,
		Payload: payload,
	}, nil
}

func (i *Implementation) GetRuleStatistic(ctx context.Context, req *desc.GetRuleRequest) (*desc.GetRuleStatisticResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetRuleStatistic")
	defer span.Finish()

	if len(req.GetService()) == 0 || len(req.GetContext()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty service or context")
	}

	stat, err := i.dillerProvider.GetRuleStatistic(ctx, req.GetService(), req.GetContext())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.GetRuleStatisticResponse{
		Scores: decodeScores(stat),
	}, nil
}

func decodeScores(in []model.Variant) []*desc.VariantScore {
	res := make([]*desc.VariantScore, len(in))

	for i, v := range in {
		res[i] = &desc.VariantScore{
			VariantId: v.Key,
			Score:     v.Score,
		}
	}

	return res
}
