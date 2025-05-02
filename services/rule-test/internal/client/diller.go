package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
)

type DillerClient interface {
	GetRuleData(ctx context.Context, req *pb.GetRuleRequest, opts ...grpc.CallOption) (*pb.GetRuleDataResponse, error)
}

type RuleDillerWrapper struct {
	cl DillerClient
}

func NewRuleDillerWrapper(cl DillerClient) *RuleDillerWrapper {
	return &RuleDillerWrapper{
		cl: cl,
	}
}

func (w *RuleDillerWrapper) GetRuleData(ctx context.Context, service, context string) (string, string, error) {
	resp, err := w.cl.GetRuleData(ctx, &pb.GetRuleRequest{Service: service, Context: context})
	if status.Code(err) == codes.Internal {
		return "", "", err
	}
	return resp.GetData(), resp.GetPayload(), nil
}
