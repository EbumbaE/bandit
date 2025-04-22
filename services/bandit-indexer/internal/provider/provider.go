package provider

import (
	"context"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) GetBandit(ctx context.Context, ruleID string) (model.Bandit, error) {
	return model.Bandit{}, nil
}
