package notifier

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

type Producer interface {
	SendMessage(ctx context.Context, key, value []byte) error
}

type Notifier struct {
	producer Producer
}

func NewNotifier(producer Producer) *Notifier {
	return &Notifier{producer: producer}
}

type Event struct {
	RuleID string `json:"rule_id"`
}

func (n *Notifier) Send(ctx context.Context, ruleID string) error {
	msg, err := json.Marshal(
		Event{
			RuleID: ruleID,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "event for rule '%s' marshal", ruleID)
	}

	return n.producer.SendMessage(ctx, []byte(ruleID), msg)
}
