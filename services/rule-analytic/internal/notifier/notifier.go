package notifier

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/rule-analytic/internal"
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

func (n *Notifier) SendEvent(ctx context.Context, event model.BanditEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return errors.Wrapf(err, "event '%v' marshal", event)
	}

	return n.producer.SendMessage(ctx, []byte(event.RuleID), msg)
}
