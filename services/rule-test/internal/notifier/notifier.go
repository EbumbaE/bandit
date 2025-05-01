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

type ActionType string

var (
	ClickActionType    = ActionType("click")
	ViewActionType     = ActionType("view")
	CartActionType     = ActionType("add_to_cart")
	PurchaseActionType = ActionType("purchase")
)

type Analitic struct {
	Payload string     `json:"payload"`
	Action  ActionType `json:"action"`
	Amount  float64    `json:"amount"`
}

func (n *Notifier) SendAnalytic(ctx context.Context, action ActionType, amount float64, payload string) error {
	event := &Analitic{
		Action:  action,
		Amount:  amount,
		Payload: payload,
	}

	msg, err := json.Marshal(event)
	if err != nil {
		return errors.Wrapf(err, "event '%v' marshal", event)
	}

	return n.producer.SendMessage(ctx, nil, msg)
}
