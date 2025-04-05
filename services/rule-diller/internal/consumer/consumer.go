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
	SaveRuleVariants(ctx context.Context, service, context string, variants []model.Variant) error
	SaveRuleVersion(ctx context.Context, service, context string, version uint64) error
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

	if err = c.storage.SaveRuleVariants(ctx, rule.Service, rule.Context, rule.Variants); err != nil {
		return errors.Wrapf(err, "SaveRuleVariants for service[%s], context[%s], variants[%v]", rule.Service, rule.Context, rule.Variants)
	}
	if err = c.storage.SaveRuleVersion(ctx, rule.Service, rule.Context, rule.Version); err != nil {
		return errors.Wrapf(err, "SaveRuleVersion for service[%s], context[%s], variants[%v]", rule.Service, rule.Context, rule.Version)
	}

	return nil
}
