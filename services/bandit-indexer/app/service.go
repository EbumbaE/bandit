package app

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/EbumbaE/bandit/pkg/genproto/bandit-indexer/api"
	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type IndexerProvider interface {
	GetBandit(ctx context.Context, ruleID string) (model.Bandit, error)
}

type Implementation struct {
	indexerProvider IndexerProvider

	desc.UnimplementedBanditIndexerServiceServer
}

func NewService(indexerProvider IndexerProvider) *Implementation {
	return &Implementation{
		indexerProvider: indexerProvider,
	}
}

func (i *Implementation) GetRuleScores(ctx context.Context, req *desc.GetRuleScoresRequest) (*desc.GetRuleScoresResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetRuleScores")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty rule id")
	}

	rule, err := i.indexerProvider.GetBandit(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.GetRuleScoresResponse{
		Version:  rule.Version,
		Variants: decodeArms(rule.Arms),
	}, nil
}

func decodeArms(in []model.Arm) []*desc.Variant {
	res := make([]*desc.Variant, len(in))

	for _, v := range in {
		res = append(res, &desc.Variant{
			Id:    v.VariantId,
			Score: v.Score,
			Count: v.Count,
		})
	}

	return res
}
