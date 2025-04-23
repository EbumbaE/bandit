package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type Provider interface {
	ApplyReward(ctx context.Context, ruleID, variantID string, reward float64, calculatedAt time.Time) error
}

type AnalyticConsumer struct {
	provider Provider
	notifier Notifier
}

func NewAnalyticConsumer(provider Provider, notifier Notifier) *AnalyticConsumer {
	return &AnalyticConsumer{
		provider: provider,
		notifier: notifier,
	}
}

type AnalyticEvent struct {
	RuleID       string    `json:"rule_id"`
	VariantID    string    `json:"variant_id"`
	Reward       float64   `json:"reward"`
	CalculatedAt time.Time `json:"calculated_at"`
}

func (c *AnalyticConsumer) Handle(ctx context.Context, msg []byte) error {
	event := &AnalyticEvent{}
	if err := json.Unmarshal(msg, event); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	if err := c.provider.ApplyReward(ctx, event.RuleID, event.VariantID, event.Reward, event.CalculatedAt); err != nil {
		return errors.Wrapf(err, "provider.ApplyReward for event [%v]", event)
	}

	return c.notifier.Send(ctx, event.RuleID)
}
