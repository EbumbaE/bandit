package provider

import (
	"context"
	"time"

	core "github.com/EbumbaE/bandit/services/bandit-core/v6"
	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type Storage interface {
	GetBanditByRuleID(ctx context.Context, ruleID string) (model.Bandit, error)
	GetArms(ctx context.Context, ruleID string) ([]model.Arm, error)
	GetArm(ctx context.Context, variantID string) (model.Arm, error)

	SetArmConfig(ctx context.Context, variantID string, config []byte) error
}

type Provider struct {
	storage Storage

	armSerializer    core.ArmSerializer
	banditSerializer core.BanditSerializer
}

func NewProvider(storage Storage) *Provider {
	return &Provider{
		storage: storage,

		armSerializer:    core.NewArmSerializer(),
		banditSerializer: core.NewBanditSerializer(),
	}
}

func (p *Provider) GetBandit(ctx context.Context, ruleID string) (model.Bandit, error) {
	bandit, err := p.storage.GetBanditByRuleID(ctx, ruleID)
	if err != nil {
		return model.Bandit{}, errors.Wrap(err, "storage.GetBanditByRuleID")
	}

	bandit.Arms, err = p.storage.GetArms(ctx, ruleID)
	if err != nil {
		return model.Bandit{}, errors.Wrap(err, "storage.GetArms")
	}

	return bandit, nil
}

func (p *Provider) ApplyReward(ctx context.Context, ruleID, variantID string, reward float64, calculatedAt time.Time) error {
	bandit, err := p.storage.GetBanditByRuleID(ctx, ruleID)
	if err != nil {
		return errors.Wrap(err, "storage.GetBanditByRuleID")
	}

	arm, err := p.storage.GetArm(ctx, variantID)
	if err != nil {
		return errors.Wrap(err, "storage.GetArm")
	}

	coreBandit, err := p.banditSerializer.Deserialize(bandit.Config)
	if err != nil {
		return errors.Wrap(err, "banditSerializer.Deserialize")
	}

	coreArm, err := p.armSerializer.Deserialize(arm.Config)
	if err != nil {
		return errors.Wrap(err, "armSerializer.Deserialize")
	}

	coreArm = coreBandit.Calculate(coreArm, reward)

	arm.Config, err = p.armSerializer.Serialize(coreArm)
	if err != nil {
		return errors.Wrap(err, "armSerializer.Serialize")
	}

	if err = p.storage.SetArmConfig(ctx, arm.VariantId, arm.Config); err != nil {
		return errors.Wrap(err, "storage.SetArmConfig")
	}

	return nil
}
