package notifier

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

type ActionType string

var (
	ActionCreate   ActionType = "create"
	ActionDelete   ActionType = "delete"
	ActionActive   ActionType = "active"
	ActionInactive ActionType = "inactive"
)

func (a ActionType) String() string {
	return string(a)
}

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
	Type   string `json:"type"`
	Action string `json:"action"`
	Id     string `json:"id"`
}

func (n *Notifier) SendRule(ctx context.Context, ruleID string, action ActionType) error {
	return n.send(ctx, ruleID, "rule", action)
}

func (n *Notifier) SendVariant(ctx context.Context, ruleID string, action ActionType) error {
	return n.send(ctx, ruleID, "variant", action)
}

func (n *Notifier) send(ctx context.Context, ruleID string, eventType string, action ActionType) error {
	msg, err := json.Marshal(
		Event{
			Type:   eventType,
			Action: action.String(),
			Id:     ruleID,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "event '%s' marshal", eventType)
	}

	return n.producer.SendMessage(ctx, []byte(ruleID), msg)
}
