package app

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	model "github.com/EbumbaE/bandit/services/rule-admin/internal"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/provider"
)

type AdminProvider interface {
	GetRule(ctx context.Context, id string) (model.Rule, error)
	CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	UpdateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	SetRuleState(ctx context.Context, id string, state model.StateType) error
	GetRuleServiceContext(ctx context.Context, ruleID string) (string, string, error)

	GetVariant(ctx context.Context, ruleID, variandID string) (model.Variant, error)
	AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error)
	SetVariantState(ctx context.Context, ruleID, variandID string, state model.StateType) error

	CreateWantedBandit(ctx context.Context, wb model.WantedBandit) error
	GetWantedRegistry(ctx context.Context) ([]model.WantedBandit, error)
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
		if errors.Is(err, provider.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeRuleResponse(r), nil
}

func (i *Implementation) CreateRule(ctx context.Context, req *desc.CreateRuleRequest) (*desc.RuleResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/CreateRule")
	defer span.Finish()

	r, err := i.ruleProvider.CreateRule(ctx, encodeCreateRule(req))
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

	if len(req.GetId()) == 0 || len(req.GetRuleId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	v, err := i.ruleProvider.GetVariant(ctx, req.GetRuleId(), req.GetId())
	if err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return decodeVariantResponse(v), nil
}

func (i *Implementation) GetVariantData(ctx context.Context, req *desc.GetVariantRequest) (*desc.VariantResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetVariant")
	defer span.Finish()

	if len(req.GetId()) == 0 || len(req.GetRuleId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	v, err := i.ruleProvider.GetVariant(ctx, req.GetRuleId(), req.GetId())
	if err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.VariantResponse{
		Variant: &desc.Variant{Data: v.Data},
	}, nil
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

	if len(req.GetId()) == 0 || len(req.GetRuleId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	if err := i.ruleProvider.SetVariantState(ctx, req.GetId(), req.GetRuleId(), encodeStateType(req.GetState())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (i *Implementation) GetRuleServiceContext(ctx context.Context, req *desc.GetRuleRequest) (*desc.GetRuleServiceContextResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetRuleServiceContext")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	service, context, err := i.ruleProvider.GetRuleServiceContext(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.GetRuleServiceContextResponse{
		Service: service,
		Context: context,
	}, nil
}

func (i *Implementation) CreateWantedBandit(ctx context.Context, req *desc.CreateWantedBanditRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/CreateWantedBandit")
	defer span.Finish()

	if len(req.GetData().GetBanditKey()) == 0 || len(req.GetData().GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty bantit key or name")
	}

	wb := model.WantedBandit{
		BanditKey: req.GetData().GetBanditKey(),
		Name:      req.GetData().GetName(),
	}

	err := i.ruleProvider.CreateWantedBandit(ctx, wb)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (i *Implementation) GetWantedRegistry(ctx context.Context, req *emptypb.Empty) (*desc.GetWantedRegistryResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/GetWantedRegistry")
	defer span.Finish()

	wr, err := i.ruleProvider.GetWantedRegistry(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &desc.GetWantedRegistryResponse{
		Registry: make([]*desc.WantedBandit, len(wr)),
	}
	for i, b := range wr {
		resp.Registry[i] = &desc.WantedBandit{
			BanditKey: b.BanditKey,
			Name:      b.Name,
		}
	}
	return resp, nil
}

func (i *Implementation) CheckRule(ctx context.Context, req *desc.CheckRequest) (*desc.CheckResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/CheckRule")
	defer span.Finish()

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	_, err := i.ruleProvider.GetRule(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return &desc.CheckResponse{IsExist: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.CheckResponse{IsExist: true}, nil
}

func (i *Implementation) CheckVariant(ctx context.Context, req *desc.CheckRequest) (*desc.CheckResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api/CheckRule")
	defer span.Finish()

	if len(req.GetId()) == 0 || len(req.GetVariantId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	_, err := i.ruleProvider.GetVariant(ctx, req.GetId(), req.GetVariantId())
	if err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return &desc.CheckResponse{IsExist: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.CheckResponse{IsExist: true}, nil
}

func encodeModifyRule(v *desc.ModifyRuleRequest) model.Rule {
	return model.Rule{
		Id:          v.Id,
		Name:        v.Name,
		Description: v.Description,
	}
}

func encodeCreateRule(v *desc.CreateRuleRequest) model.Rule {
	return model.Rule{
		Name:        v.Name,
		Description: v.Description,
		Variants:    encodeVariants(v.GetVariants()),
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
		BanditKey:   r.BanditKey,
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
		Name:  v.Name,
		Data:  v.Data,
		State: decodeStateType(v.State),
	}
}

func encodeVariant(v *desc.Variant) model.Variant {
	return model.Variant{
		Id:    v.GetId(),
		Name:  v.GetName(),
		Data:  v.GetData(),
		State: encodeStateType(v.State),
	}
}

func encodeVariants(in []*desc.Variant) []model.Variant {
	out := make([]model.Variant, len(in))
	for i, v := range in {
		out[i] = encodeVariant(v)
	}
	return out
}

func decodeVariantResponse(v model.Variant) *desc.VariantResponse {
	return &desc.VariantResponse{
		Variant: decodeVariant(v),
	}
}
