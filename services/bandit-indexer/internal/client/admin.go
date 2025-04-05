package wrapper

import (
	"context"

	pb "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"google.golang.org/grpc"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type AdminClient interface {
	GetRule(ctx context.Context, in *pb.GetRuleRequest, opts ...grpc.CallOption) (*pb.Rule, error)
	GetVariant(ctx context.Context, in *pb.GetVariantRequest, opts ...grpc.CallOption) (*pb.Variant, error)
}

type AdminWrapper struct {
	client AdminClient
}

func NewAdminWrapper(client AdminClient) *AdminWrapper {
	return &AdminWrapper{
		client: client,
	}
}

func (i *AdminWrapper) GetRule(ctx context.Context, ruleID string) (model.Rule, error) {
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
