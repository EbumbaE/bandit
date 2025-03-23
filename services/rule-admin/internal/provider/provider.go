package provider

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/logger"
	model "github.com/EbumbaE/bandit/services/rule-admin/internal"
)

var ErrNotFound = errors.New("not found")

type Storage interface {
	GetRule(ctx context.Context, id string) (model.Rule, error)
	CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	UpdateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	SetRuleState(ctx context.Context, id string, state model.StateType) error

	GetVariant(ctx context.Context, ruleID, variantID string) (model.Variant, error)
	GetVariants(ctx context.Context, ruleID string) ([]model.Variant, error)
	AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error)
	SetVariantState(ctx context.Context, id string, state model.StateType) error
}

type Provider struct {
	storage Storage
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) GetRule(ctx context.Context, id string) (model.Rule, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/GetRule")
	defer span.Finish()

	r, err := p.GetRule(ctx, id)
	if err != nil {
		return model.Rule{}, err
	}

	r.Variants, err = p.storage.GetVariants(ctx, id)
	if err != nil {
		span.SetTag("error", err)
		logger.Error("provider/GetRule: GetVariants", zap.Error(err))
	}

	return r, nil
}

func (p *Provider) CreateRule(ctx context.Context, r model.Rule) (model.Rule, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/CreateRule")
	defer span.Finish()

	r, err := p.storage.CreateRule(ctx, r)
	if err != nil {
		return model.Rule{}, err
	}

	var createdVariants []model.Variant
	for _, v := range r.Variants {
		v, err = p.AddVariant(ctx, r.Id, v)
		if err != nil {
			return model.Rule{}, err
		}
		createdVariants = append(createdVariants, v)
	}
	r.Variants = createdVariants

	return r, nil
}

func (p *Provider) UpdateRule(ctx context.Context, r model.Rule) (model.Rule, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/UpdateRule")
	defer span.Finish()

	r, err := p.storage.UpdateRule(ctx, r)
	if err != nil {
		return model.Rule{}, err
	}

	return r, nil
}

func (p *Provider) SetRuleState(ctx context.Context, id string, state model.StateType) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/SetRuleState")
	defer span.Finish()

	if err := p.storage.SetRuleState(ctx, id, state); err != nil {
		return err
	}

	return nil
}

func (p *Provider) GetVariant(ctx context.Context, ruleID, variandID string) (model.Variant, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/GetVariant")
	defer span.Finish()

	v, err := p.storage.GetVariant(ctx, ruleID, variandID)
	if err != nil {
		return model.Variant{}, err
	}

	return v, nil
}

func (p *Provider) AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/AddVariant")
	defer span.Finish()

	v, err := p.storage.AddVariant(ctx, ruleID, v)
	if err != nil {
		return model.Variant{}, err
	}

	return v, nil
}

func (p *Provider) SetVariantState(ctx context.Context, id string, state model.StateType) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/SetVariantState")
	defer span.Finish()

	if err := p.storage.SetVariantState(ctx, id, state); err != nil {
		return err
	}

	return nil
}
