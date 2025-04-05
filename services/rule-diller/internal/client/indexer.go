package wrapper

import (
	"context"

	pb "github.com/EbumbaE/bandit/pkg/genproto/bandit-indexer/api"
	"google.golang.org/grpc"

	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
)

type IndexerClient interface {
	GetRuleScores(ctx context.Context, in *pb.GetRuleScoresRequest, opts ...grpc.CallOption) (*pb.GetRuleScoresResponse, error)
}

type IndexerWrapper struct {
	client IndexerClient
}

func NewIndexerWrapper(client IndexerClient) *IndexerWrapper {
	return &IndexerWrapper{
		client: client,
	}
}

func (i *IndexerWrapper) GetRule(ctx context.Context, ruleID string) (model.Rule, error) {
	rule, err := i.client.GetRuleScores(ctx, &pb.GetRuleScoresRequest{Id: ruleID})
	if err != nil {
		return model.Rule{}, err
	}

	return model.Rule{
		Service:  rule.GetService(),
		Context:  rule.GetContext(),
		Variants: decodeVariants(rule.GetVariants()),
		Version:  rule.GetVersion(),
	}, nil
}

func decodeVariants(in []*pb.Variant) []model.Variant {
	res := make([]model.Variant, len(in))

	for i, v := range in {
		res[i] = model.Variant{
			Key:   v.GetId(),
			Data:  v.GetData(),
			Score: v.GetScore(),
			Count: v.GetCount(),
		}
	}

	return res
}
