package app

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
)

type DillerProvider interface {
	GetRuleData(ctx context.Context, service, ctxKey string) ([]byte, error)
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

	ruleData, err := i.dillerProvider.GetRuleData(ctx, req.GetService(), req.GetContext())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.GetRuleDataResponse{
		Data: ruleData,
	}, nil
}

func (i *Implementation) GetRuleStatistic(ctx context.Context, req *desc.GetRuleRequest) (*desc.GetRuleStatisticResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetRule")
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

	for _, v := range in {
		res = append(res, &desc.VariantScore{
			VariantId: v.Key,
			Score:     v.Score,
			Data:      v.Data,
		})
	}

	return res
}
