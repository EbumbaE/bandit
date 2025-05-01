package wrapper

import (
	"context"

	pb "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"google.golang.org/grpc"
)

type AdminClient interface {
	GetRuleServiceContext(ctx context.Context, in *pb.GetRuleRequest, opts ...grpc.CallOption) (*pb.GetRuleServiceContextResponse, error)
	GetVariantData(ctx context.Context, in *pb.GetVariantRequest, opts ...grpc.CallOption) (*pb.VariantResponse, error)
}

type AdminWrapper struct {
	client AdminClient
}

func NewAdminWrapper(client AdminClient) *AdminWrapper {
	return &AdminWrapper{
		client: client,
	}
}

func (i *AdminWrapper) GetRuleServiceContext(ctx context.Context, ruleID string) (string, string, error) {
	resp, err := i.client.GetRuleServiceContext(ctx, &pb.GetRuleRequest{Id: ruleID})
	return resp.GetService(), resp.GetContext(), err
}

func (i *AdminWrapper) GetVariantData(ctx context.Context, ruleID string, variantID string) (string, error) {
	resp, err := i.client.GetVariantData(ctx, &pb.GetVariantRequest{Id: variantID, RuleId: ruleID})
	return resp.GetVariant().GetData(), err
}
