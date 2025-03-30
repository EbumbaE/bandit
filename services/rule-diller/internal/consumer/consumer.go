package consumer

import (
	"context"
	"encoding/json"

	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
	"github.com/pkg/errors"
)

type Indexer interface {
	GetRule(ctx context.Context, ruleID string) (model.Rule, error)
}

type Storage interface {
	SaveRuleVariants(ctx context.Context, key string, variants []model.Variant) error
	SaveRuleVersion(ctx context.Context, key string, version uint64) error

	SaveVariantData(ctx context.Context, key string, data []byte) error
	SetVariantCount(ctx context.Context, key string, count uint64) error
}

type Consumer struct {
	storage Storage
	indexer Indexer
}

func NewConsumer(indexer Indexer, storage Storage) *Consumer {
	return &Consumer{
		indexer: indexer,
		storage: storage,
	}
}

type Event struct {
	RuleID string `json:"rule_id"`
}

func (c *Consumer) Handle(ctx context.Context, msg []byte) error {
	event := &Event{}
	if err := json.Unmarshal(msg, event); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	rule, err := c.indexer.GetRule(ctx, event.RuleID)
	if err != nil {
		return errors.Wrapf(err, "GetRule for rule[%s]", event.RuleID)
	}

	ruleKey := model.RuleKey{
		Service: rule.Service,
		Context: rule.Context,
	}.GetKey()

	if err = c.storage.SaveRuleVariants(ctx, ruleKey, rule.Variants); err != nil {
		return errors.Wrapf(err, "SaveRuleVariants for ruleKey[%s], variants[%v]", ruleKey, rule.Variants)
	}
	if err = c.storage.SaveRuleVersion(ctx, ruleKey, rule.Version); err != nil {
		return errors.Wrapf(err, "SaveRuleVersion for ruleKey[%s], variants[%v]", ruleKey, rule.Version)
	}

	for _, v := range rule.Variants {
		if err = c.storage.SaveVariantData(ctx, ruleKey, v.Data); err != nil {
			return errors.Wrapf(err, "SaveVariantData for variant[%v]", v)
		}
		if err = c.storage.SetVariantCount(ctx, ruleKey, v.Count); err != nil {
			return errors.Wrapf(err, "SetVariantCount for variant[%v]", v)
		}
	}

	return nil
}
