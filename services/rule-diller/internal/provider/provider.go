package provider

import (
	"context"

	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) GetRuleData(ctx context.Context, service, ctxKey string) ([]byte, error) {
	return nil, nil
}

func (p *Provider) GetRuleStatistic(ctx context.Context, service, ctxKey string) ([]model.Variant, error) {
	return nil, nil
}
