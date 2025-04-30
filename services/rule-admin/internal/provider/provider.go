package provider

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/logger"
	model "github.com/EbumbaE/bandit/services/rule-admin/internal"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/notifier"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/storage"
)

var ErrNotFound = errors.New("not found")

type Storage interface {
	GetRule(ctx context.Context, id string) (model.Rule, error)
	CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	UpdateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	SetRuleState(ctx context.Context, id string, state model.StateType) error
	GetRuleServiceContext(ctx context.Context, ruleID string) (string, string, error)
	GetActiveRuleByServiceContext(ctx context.Context, service, context string) (string, error)

	GetVariant(ctx context.Context, ruleID, variantID string) (model.Variant, error)
	GetVariants(ctx context.Context, ruleID string) ([]model.Variant, error)
	AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error)
	SetVariantState(ctx context.Context, id string, state model.StateType) error

	CreateWantedBandit(ctx context.Context, wb model.WantedBandit) error
	GetWantedRegistry(ctx context.Context) ([]model.WantedBandit, error)
	CheckWantedBandit(ctx context.Context, banditKey string) (bool, error)
}

type Notifier interface {
	SendRule(ctx context.Context, ruleID string, action notifier.ActionType) error
	SendVariant(ctx context.Context, ruleID, variantID string, action notifier.ActionType) error
}

type Provider struct {
	storage  Storage
	notifier Notifier
}

func NewProvider(storage Storage, notifier Notifier) *Provider {
	return &Provider{
		storage:  storage,
		notifier: notifier,
	}
}

func (p *Provider) GetRule(ctx context.Context, id string) (model.Rule, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/GetRule")
	defer span.Finish()

	r, err := p.storage.GetRule(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return model.Rule{}, ErrNotFound
		}
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

	ok, err := p.storage.CheckWantedBandit(ctx, r.BanditKey)
	if err != nil {
		return model.Rule{}, err
	}
	if !ok {
		return model.Rule{}, errors.New("validate bandit key")
	}

	id, err := p.storage.GetActiveRuleByServiceContext(ctx, r.Service, r.Context)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return model.Rule{}, err
	}
	if len(id) > 0 {
		return model.Rule{}, fmt.Errorf("active rule already exist[%s]", id)
	}

	variants := r.Variants

	r, err = p.storage.CreateRule(ctx, r)
	if err != nil {
		return model.Rule{}, err
	}

	for _, v := range variants {
		addedV, err := p.AddVariant(ctx, r.Id, v)
		if err != nil {
			return r, err
		}
		r.Variants = append(r.Variants, addedV)
	}

	if err := p.notifier.SendRule(ctx, r.Id, notifier.ActionCreate); err != nil {
		logger.Error("failed send create rule event", zap.Error(err))
	}

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

	switch state {
	case model.StateTypeDisable:
		if err := p.notifier.SendRule(ctx, id, notifier.ActionInactive); err != nil {
			logger.Error("failed send inactive rule event", zap.Error(err))
		}
	case model.StateTypeEnable:
		if err := p.notifier.SendRule(ctx, id, notifier.ActionActive); err != nil {
			logger.Error("failed send active rule event", zap.Error(err))
		}
	}

	return nil
}

func (p *Provider) GetVariant(ctx context.Context, ruleID, variandID string) (model.Variant, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/GetVariant")
	defer span.Finish()

	v, err := p.storage.GetVariant(ctx, ruleID, variandID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return model.Variant{}, ErrNotFound
		}
		return model.Variant{}, err
	}

	return v, nil
}

func (p *Provider) AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/AddVariant")
	defer span.Finish()

	_, err := p.storage.GetRule(ctx, ruleID)
	if err != nil {
		return model.Variant{}, err
	}

	v, err = p.storage.AddVariant(ctx, ruleID, v)
	if err != nil {
		return model.Variant{}, err
	}

	if err := p.notifier.SendVariant(ctx, ruleID, v.Id, notifier.ActionCreate); err != nil {
		logger.Error("failed send create variant event", zap.Error(err))
	}

	return v, nil
}

func (p *Provider) SetVariantState(ctx context.Context, ruleID, variantID string, state model.StateType) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "provider/SetVariantState")
	defer span.Finish()

	if err := p.storage.SetVariantState(ctx, variantID, state); err != nil {
		return err
	}

	switch state {
	case model.StateTypeDisable:
		if err := p.notifier.SendVariant(ctx, ruleID, variantID, notifier.ActionCreate); err != nil {
			logger.Error("failed send inactive variant event", zap.Error(err))
		}
	case model.StateTypeEnable:
		if err := p.notifier.SendVariant(ctx, ruleID, variantID, notifier.ActionCreate); err != nil {
			logger.Error("failed send active variant event", zap.Error(err))
		}
	}

	return nil
}

func (p *Provider) CreateWantedBandit(ctx context.Context, wb model.WantedBandit) error {
	return p.storage.CreateWantedBandit(ctx, wb)
}

func (p *Provider) GetWantedRegistry(ctx context.Context) ([]model.WantedBandit, error) {
	return p.storage.GetWantedRegistry(ctx)
}

func (p *Provider) GetRuleServiceContext(ctx context.Context, ruleID string) (string, string, error) {
	return p.storage.GetRuleServiceContext(ctx, ruleID)
}
