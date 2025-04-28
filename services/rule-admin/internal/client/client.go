package client

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
)

type Client interface {
	GetRuleStatistic(ctx context.Context, req *pb.GetRuleRequest, opts ...grpc.CallOption) (*pb.GetRuleStatisticResponse, error)
}

type RuleDillerWrapper struct {
	cl Client
}

func NewRuleDillerWrapper(cl Client) *RuleDillerWrapper {
	return &RuleDillerWrapper{
		cl: cl,
	}
}

func (w *RuleDillerWrapper) GetRuleStatistic(ctx context.Context, service, context string) (map[string]float64, error) {
	stat, err := w.cl.GetRuleStatistic(ctx, &pb.GetRuleRequest{Service: service, Context: context})
	if err != nil {
		return nil, err
	}

	res := map[string]float64{}
	for _, score := range stat.GetScores() {
		res[score.GetVariantId()] = score.Score
	}

	return res, nil
}
