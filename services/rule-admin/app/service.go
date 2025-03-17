package token

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	model "github.com/EbumbaE/bandit/services/rule-admin/internal"
)

type AdminProvider interface {
	GetRule(ctx context.Context, id string) (model.Rule, error)
	CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	UpdateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	SetRuleState(ctx context.Context, id string, state model.StateType) error

	GetVariant(ctx context.Context, id string) (model.Variant, error)
	AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error)
	SetVariantState(ctx context.Context, id string, state model.StateType) error
}

type Implementation struct {
	ruleProvider AdminProvider

	desc.UnimplementedRuleAdminServiceServer
}

func NewService(ruleProvider AdminProvider) *Implementation {
	return &Implementation{
		ruleProvider: ruleProvider,
	}
}

func (i *Implementation) GetRule(ctx context.Context, req *desc.GetRuleRequest) (*desc.RuleResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetRule")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	r, err := i.ruleProvider.GetRule(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeRuleResponse(r), nil
}

func (i *Implementation) CreateRule(ctx context.Context, req *desc.ModifyRuleRequest) (*desc.RuleResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/CreateRule")
	defer span.Finish()

	r, err := i.ruleProvider.CreateRule(ctx, encodeModifyRule(req))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeRuleResponse(r), nil
}

func (i *Implementation) UpdateRule(ctx context.Context, req *desc.ModifyRuleRequest) (*desc.RuleResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/UpdateRule")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	r, err := i.ruleProvider.UpdateRule(ctx, encodeModifyRule(req))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeRuleResponse(r), nil
}

func (i *Implementation) SetRuleState(ctx context.Context, req *desc.SetRuleStateRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/SetRuleState")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	if err := i.ruleProvider.SetRuleState(ctx, req.GetId(), encodeStateType(req.GetState())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (i *Implementation) GetVariant(ctx context.Context, req *desc.GetVariantRequest) (*desc.VariantResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetVariant")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	v, err := i.ruleProvider.GetVariant(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeVariantResponse(v), nil
}

func (i *Implementation) AddVariant(ctx context.Context, req *desc.AddVariantRequest) (*desc.VariantResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/AddVariant")
	defer span.Finish()

	if len(req.GetRuleId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty rule id")
	}

	r, err := i.ruleProvider.AddVariant(ctx, req.GetRuleId(), encodeVariant(req.GetVariant()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeVariantResponse(r), nil
}

func (i *Implementation) SetVariantState(ctx context.Context, req *desc.SetVariantStateRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/SetVariantState")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	if err := i.ruleProvider.SetVariantState(ctx, req.GetId(), encodeStateType(req.GetState())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func encodeModifyRule(v *desc.ModifyRuleRequest) model.Rule {
	return model.Rule{
		Id:          v.Id,
		Name:        v.Name,
		Description: v.Description,
	}
}

func decodeRuleResponse(r model.Rule) *desc.RuleResponse {
	return &desc.RuleResponse{
		Rule: decodeRule(r),
	}
}

func decodeRule(r model.Rule) *desc.Rule {
	return &desc.Rule{
		Id:          r.Id,
		Name:        r.Name,
		Description: r.Description,
		State:       decodeStateType(r.State),
		Variants:    decodeVariants(r.Variants),
	}
}

func decodeStateType(v model.StateType) desc.State {
	switch v {
	case model.StateTypeEnable:
		return desc.State_RULE_STATE_ENABLED
	case model.StateTypeDisable:
		return desc.State_RULE_STATE_DISABLED
	default:
		return desc.State_RULE_STATE_UNSPECIFIED
	}
}

func encodeStateType(v desc.State) model.StateType {
	switch v {
	case desc.State_RULE_STATE_ENABLED:
		return model.StateTypeEnable
	default:
		return model.StateTypeDisable
	}
}

func decodeVariants(in []model.Variant) []*desc.Variant {
	out := make([]*desc.Variant, len(in))
	for i, v := range in {
		out[i] = decodeVariant(v)
	}
	return out
}

func decodeVariant(v model.Variant) *desc.Variant {
	return &desc.Variant{
		Id:    v.Id,
		Data:  nil,
		State: decodeStateType(v.State),
	}
}

func encodeVariant(v *desc.Variant) model.Variant {
	d := v.GetData()
	marshaled, _ := json.Marshal(d)

	return model.Variant{
		Id:    v.GetId(),
		Data:  marshaled,
		State: encodeStateType(v.State),
	}
}

func decodeVariantResponse(v model.Variant) *desc.VariantResponse {
	return &desc.VariantResponse{
		Variant: decodeVariant(v),
	}
}
