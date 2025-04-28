package provider

import (
	"context"

	core "github.com/EbumbaE/bandit/services/bandit-core/v6"
	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type Storage interface {
	GetBanditByRuleID(ctx context.Context, ruleID string) (model.Bandit, error)
	GetArms(ctx context.Context, ruleID string) ([]model.Arm, error)
	GetArm(ctx context.Context, variantID string) (model.Arm, error)

	SetArmConfig(ctx context.Context, variantID string, config []byte) error
	UpBanditVersion(ctx context.Context, ruleID string) error
}

type Provider struct {
	storage Storage
}

func NewProvider(storage Storage) *Provider {
	return &Provider{
		storage: storage,
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

func (p *Provider) ApplyReward(ctx context.Context, ruleID, variantID string, reward float64, version uint64) error {
	bandit, err := p.storage.GetBanditByRuleID(ctx, ruleID)
	if err != nil {
		return errors.Wrap(err, "storage.GetBanditByRuleID")
	}

	arm, err := p.storage.GetArm(ctx, variantID)
	if err != nil {
		return errors.Wrap(err, "storage.GetArm")
	}

	coreBandit := core.NewDefaultGaussianBandit()
	if err := coreBandit.Deserialize(bandit.Config); err != nil {
		return errors.Wrap(err, "coreBandit.Deserialize")
	}
	coreBandit.Version = bandit.Version

	coreArm := core.NewDefaultGaussianArm()
	if err := coreArm.Deserialize(arm.Config); err != nil {
		return errors.Wrap(err, "coreArm.Deserialize")
	}
	coreArm.Version = version

	coreArm = coreBandit.Calculate(coreArm, reward)

	arm.Config, err = coreArm.Serialize()
	if err != nil {
		return errors.Wrap(err, "coreArm.Serialize")
	}

	if err = p.storage.SetArmConfig(ctx, arm.VariantId, arm.Config); err != nil {
		return errors.Wrap(err, "storage.SetArmConfig")
	}

	if err = p.storage.UpBanditVersion(ctx, bandit.RuleId); err != nil {
		return errors.Wrap(err, "storage.UpBanditVersion")
	}

	return nil
}
