package client

import (
	"context"

	pb "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"github.com/EbumbaE/bandit/services/bandit-core/v6"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AdminClient interface {
	CreateRule(ctx context.Context, req *pb.CreateRuleRequest, opts ...grpc.CallOption) (*pb.RuleResponse, error)

	AddVariant(ctx context.Context, req *pb.AddVariantRequest, opts ...grpc.CallOption) (*pb.VariantResponse, error)
	SetVariantState(ctx context.Context, req *pb.SetVariantStateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)

	CreateWantedBandit(ctx context.Context, req *pb.CreateWantedBanditRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetWantedRegistry(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetWantedRegistryResponse, error)
}

type RuleAdminWrapper struct {
	cl AdminClient
}

func NewRuleAdminWrapper(cl AdminClient) *RuleAdminWrapper {
	return &RuleAdminWrapper{
		cl: cl,
	}
}

func (w *RuleAdminWrapper) CreateRule(ctx context.Context, service, context string) (string, error) {
	resp, err := w.cl.CreateRule(ctx, &pb.CreateRuleRequest{
		Name:        "test",
		Description: "for test",
		Service:     service,
		Context:     context,
		BanditKey:   "gaussian",
		Variants: []*pb.Variant{
			{
				Name:  "arm1",
				Data:  `{"Value": [1, 2, 3]}`,
				State: pb.State_STATE_ENABLED,
			},
			{
				Name:  "arm2",
				Data:  `{"Value": [4, 5, 6]}`,
				State: pb.State_STATE_ENABLED,
			},
		},
	})

	return resp.GetRule().GetId(), err
}

func (w *RuleAdminWrapper) AddVariant(ctx context.Context, ruleID string) (string, error) {
	resp, err := w.cl.AddVariant(ctx, &pb.AddVariantRequest{
		RuleId: ruleID,
		Variant: &pb.Variant{
			Name:  "arm3",
			Data:  `{"Value": [7, 8, 9]}`,
			State: pb.State_STATE_ENABLED,
		},
	})
	return resp.GetVariant().GetId(), err
}

func (w *RuleAdminWrapper) DisableVariant(ctx context.Context, ruleID, variantID string) error {
	_, err := w.cl.SetVariantState(ctx, &pb.SetVariantStateRequest{
		Id:     variantID,
		RuleId: ruleID,
		State:  pb.State_STATE_DISABLED,
	})
	return err
}

func (w *RuleAdminWrapper) CreateGaussianBanditIfExist(ctx context.Context) error {
	registry, err := w.cl.GetWantedRegistry(ctx, nil)
	if err != nil {
		return err
	}

	notExist := true
	for _, w := range registry.GetRegistry() {
		if w.GetBanditKey() == bandit.GaussianBanditKey {
			notExist = false
			break
		}
	}

	if notExist {
		_, err = w.cl.CreateWantedBandit(ctx, &pb.CreateWantedBanditRequest{
			Data: &pb.WantedBandit{
				BanditKey: bandit.GaussianBanditKey,
				Name:      "Бандит Гаусса",
			},
		})
	}

	return err
}
