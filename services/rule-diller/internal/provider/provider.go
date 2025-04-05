package provider

import (
	"context"
	"encoding/json"

	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/services/bandit-core/v5"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
)

type Storage interface {
	GetRuleVariants(ctx context.Context, service, context string, withData bool) ([]model.Variant, error)
	GetRuleVersion(ctx context.Context, service, context string) (uint64, error)
	GetVariantData(ctx context.Context, service, context, variantID string) ([]byte, error)

	IncVariantCount(ctx context.Context, service, context, variantID string) error
}

type Provider struct {
	storage Storage
}

func NewProvider(storage Storage) *Provider {
	return &Provider{
		storage: storage,
	}
}

func (p *Provider) GetRuleData(ctx context.Context, service, ctxKey string) ([]byte, []byte, error) {
	variants, err := p.storage.GetRuleVariants(ctx, service, ctxKey, false)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "GetRuleVariants for service[%s], context[%s]", service, ctxKey)
	}

	options := convertToProperties(variants)
	selectedKey := bandit.SelectByProbabilities(options, bandit.DefaultExplorationFactor)

	if err := p.storage.IncVariantCount(ctx, service, ctxKey, selectedKey); err != nil {
		logger.Error("IncVariantCount", zap.String("variant_key", selectedKey), zap.Error(err))
	}

	version, err := p.storage.GetRuleVersion(ctx, service, ctxKey)
	if err != nil {
		logger.Error("GetRuleVersion", zap.String("variant_key", selectedKey), zap.Error(err))
	}

	data, err := p.storage.GetVariantData(ctx, service, ctxKey, selectedKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "GetVariantData for variant[%s]", selectedKey)
	}

	payload, err := json.Marshal(model.PayloadAnalitic{
		Service:       service,
		Context:       ctxKey,
		VariantID:     selectedKey,
		BanditVersion: version,
	})
	if err != nil {
		logger.Error("json marshal payload", zap.String("variant_key", selectedKey), zap.Error(err))
	}

	return data, payload, nil
}

func convertToProperties(variants []model.Variant) map[string]bandit.Probability {
	result := make(map[string]bandit.Probability, len(variants))

	for _, v := range variants {
		result[v.Key] = bandit.Probability{
			Score: v.Score,
			Count: v.Count,
		}
	}

	return result
}

func (p *Provider) GetRuleStatistic(ctx context.Context, service, ctxKey string) ([]model.Variant, error) {
	variants, err := p.storage.GetRuleVariants(ctx, service, ctxKey, true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetRuleVariants for service[%s], context[%s]", service, ctxKey)
	}

	return variants, nil
}
