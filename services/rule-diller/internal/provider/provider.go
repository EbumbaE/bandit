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
	GetRuleVariants(ctx context.Context, key string) ([]model.Variant, error)
	GetRuleVersion(ctx context.Context, key string) (uint64, error)

	GetVariantData(ctx context.Context, key string) ([]byte, error)
	GetVariantCount(ctx context.Context, key string) (uint64, error)
	IncVariantCount(ctx context.Context, key string) error
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
	key := model.RuleKey{
		Service: service,
		Context: ctxKey,
	}.GetKey()

	variants, err := p.storage.GetRuleVariants(ctx, key)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "GetRuleVariants for key[%s]", key)
	}

	for i, v := range variants {
		variants[i].Count, err = p.storage.GetVariantCount(ctx, v.Key)
		if err != nil {
			logger.Error("GetVariantCount", zap.String("variant_key", v.Key), zap.Error(err))
		}
	}

	options := convertToProperties(variants)
	selectedKey := bandit.SelectByProbabilities(options, bandit.DefaultExplorationFactor)

	if err := p.storage.IncVariantCount(ctx, selectedKey); err != nil {
		logger.Error("IncVariantCount", zap.String("variant_key", selectedKey), zap.Error(err))
	}

	version, err := p.storage.GetRuleVersion(ctx, selectedKey)
	if err != nil {
		logger.Error("GetRuleVersion", zap.String("variant_key", selectedKey), zap.Error(err))
	}

	data, err := p.storage.GetVariantData(ctx, selectedKey)
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
	key := model.RuleKey{
		Service: service,
		Context: ctxKey,
	}.GetKey()

	variants, err := p.storage.GetRuleVariants(ctx, key)
	if err != nil {
		return nil, errors.Wrapf(err, "GetRuleVariants for key[%s]", key)
	}

	for i, v := range variants {
		variants[i].Count, err = p.storage.GetVariantCount(ctx, v.Key)
		if err != nil {
			logger.Error("GetVariantCount", zap.String("variant_key", v.Key), zap.Error(err))
		}

		variants[i].Data, err = p.storage.GetVariantData(ctx, v.Key)
		if err != nil {
			logger.Error("GetVariantData", zap.String("variant_key", v.Key), zap.Error(err))
		}
	}

	return variants, nil
}
