package wrapper

import (
	"context"

	pb "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"google.golang.org/grpc"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type AdminClient interface {
	GetRule(ctx context.Context, in *pb.GetRuleRequest, opts ...grpc.CallOption) (*pb.RuleResponse, error)
	CheckRule(ctx context.Context, in *pb.CheckRequest, opts ...grpc.CallOption) (*pb.CheckResponse, error)

	GetVariant(ctx context.Context, in *pb.GetVariantRequest, opts ...grpc.CallOption) (*pb.VariantResponse, error)
	CheckVariant(ctx context.Context, in *pb.CheckRequest, opts ...grpc.CallOption) (*pb.CheckResponse, error)
}

type AdminWrapper struct {
	client AdminClient
}

func NewAdminWrapper(client AdminClient) *AdminWrapper {
	return &AdminWrapper{
		client: client,
	}
}

func (i *AdminWrapper) GetBandit(ctx context.Context, ruleID string) (model.Bandit, error) {
	rule, err := i.client.GetRule(ctx, &pb.GetRuleRequest{Id: ruleID})
	if err != nil {
		return model.Bandit{}, err
	}

	return model.Bandit{
		RuleId:    rule.GetRule().GetId(),
		BanditKey: rule.GetRule().GetBanditKey(),
		State:     decodeStateType(rule.GetRule().GetState()),
		Arms:      decodeArms(rule.GetRule().GetVariants()),
	}, nil
}

func (i *AdminWrapper) CheckBandit(ctx context.Context, ruleID string) (bool, error) {
	check, err := i.client.CheckRule(ctx, &pb.CheckRequest{Id: ruleID})
	return check.GetIsExist(), err
}

func (i *AdminWrapper) GetBanditState(ctx context.Context, ruleID string) (model.StateType, error) {
	rule, err := i.client.GetRule(ctx, &pb.GetRuleRequest{Id: ruleID})
	return decodeStateType(rule.GetRule().GetState()), err
}

func (i *AdminWrapper) CheckArm(ctx context.Context, ruleID, variantID string) (bool, error) {
	check, err := i.client.CheckVariant(ctx, &pb.CheckRequest{Id: ruleID, VariantId: variantID})
	return check.GetIsExist(), err
}

func (i *AdminWrapper) GetArm(ctx context.Context, ruleID, variantID string) (model.Arm, error) {
	variant, err := i.client.GetVariant(ctx, &pb.GetVariantRequest{Id: variantID, RuleId: ruleID})
	return decodeArm(variant.GetVariant()), err
}

func (i *AdminWrapper) GetArmState(ctx context.Context, ruleID, variantID string) (model.StateType, error) {
	rule, err := i.client.GetVariant(ctx, &pb.GetVariantRequest{Id: variantID, RuleId: ruleID})
	return decodeStateType(rule.GetVariant().GetState()), err
}

func decodeArms(in []*pb.Variant) []model.Arm {
	res := make([]model.Arm, len(in))

	for i, v := range in {
		res[i] = decodeArm(v)
	}

	return res
}

func decodeArm(v *pb.Variant) model.Arm {
	return model.Arm{
		VariantId: v.GetId(),
		State:     decodeStateType(v.GetState()),
	}
}

func decodeStateType(state pb.State) model.StateType {
	switch state {
	case pb.State_STATE_ENABLED:
		return model.StateTypeEnable
	case pb.State_STATE_DISABLED:
		return model.StateTypeDisable
	default:
		return model.StateTypeDisable
	}
}
