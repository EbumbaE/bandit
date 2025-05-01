package consumer

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/rule-analytic/internal"
)

type Storage interface {
	CreateAnalyticEvent(ctx context.Context, event model.BanditEvent) error
	RemoveAnalyticEvent(ctx context.Context, ruleID, variantID string) error
	InsertHistoryBatch(ctx context.Context, batch []model.HistoryEvent) error
}

type Notifier interface {
	SendEvent(ctx context.Context, event model.BanditEvent) error
}

type Consumer struct {
	storage  Storage
	notifier Notifier
}

func NewConsumer(storage Storage, notifier Notifier) *Consumer {
	return &Consumer{
		storage:  storage,
		notifier: notifier,
	}
}

type Event struct {
	Payload string           `json:"payload"`
	Action  model.ActionType `json:"action"`
	Amount  float64          `json:"amount"`
}

func (c *Consumer) Handle(ctx context.Context, msg []byte) error {
	unmarhaled := &Event{}
	if err := json.Unmarshal(msg, unmarhaled); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	if unmarhaled.Payload == "" {
		return nil
	}

	toHistory := model.HistoryEvent{
		Action: unmarhaled.Action,
		Amount: unmarhaled.Amount,
	}
	if err := json.Unmarshal([]byte(unmarhaled.Payload), &(toHistory.Payload)); err != nil {
		return errors.Wrapf(err, "unmarshal message: %s", string(msg))
	}

	reward := toHistory.Amount

	switch toHistory.Action {
	case model.ClickActionType:
		reward *= 0.3
	case model.ViewActionType:
		reward *= 0.1
	case model.CartActionType:
		reward *= 0.003 // 100 у.e == 1 клик
	case model.PurchaseActionType:
		reward *= 0.3 // 1 у.e == 1 клик
	default:
		reward *= 0
	}

	toSend := model.BanditEvent{
		RuleID:      toHistory.Payload.RuleID,
		VariantID:   toHistory.Payload.VariantID,
		Reward:      reward,
		RuleVersion: toHistory.Payload.RuleVersion,
	}

	// TODO асинхронная отправка ивента
	// if err := c.storage.CreateAnalyticEvent(ctx, toSend); err != nil {
	// 	return errors.Wrapf(err, "storage.CreateAnalyticEvent [%v]", toSend)
	// }

	if err := c.notifier.SendEvent(ctx, toSend); err != nil {
		return errors.Wrapf(err, "notifier.SendEvent [%v]", toSend)
	}

	// TODO нормальная запись батчами
	if err := c.storage.InsertHistoryBatch(ctx, []model.HistoryEvent{toHistory}); err != nil {
		return errors.Wrapf(err, "storage.InsertHistoryBatch [%v]", toSend)
	}

	return nil
}
